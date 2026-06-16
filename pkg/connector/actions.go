package connector

import (
	"context"
	"errors"
	"fmt"

	configv1 "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/conductorone/baton-retool/pkg/client"
)

const (
	ActionEnableUser  = "enable_user"
	ActionDisableUser = "disable_user"
)

// Retool's REST API has no hard delete: DELETE /api/v2/users/{id} merely sets
// active=false. The account lifecycle after create is therefore modeled as the
// reversible enable/disable pair below (PATCH /active) instead of a Delete that
// would misrepresent a deactivation as a removal.
var enableUserAction = v2.BatonActionSchema_builder{
	Name:        ActionEnableUser,
	DisplayName: "Enable User",
	Description: "Enables (reactivates) a disabled Retool user account, restoring sign-in.",
	ActionType:  []v2.ActionType{v2.ActionType_ACTION_TYPE_ACCOUNT_ENABLE},
	Arguments: []*configv1.Field{
		configv1.Field_builder{
			Name:        "user_id",
			DisplayName: "User",
			Description: "The Retool user to enable (reactivate).",
			IsRequired:  true,
			ResourceIdField: configv1.ResourceIdField_builder{
				Rules: configv1.ResourceIDRules_builder{
					AllowedResourceTypeIds: []string{resourceTypeUser.Id},
				}.Build(),
			}.Build(),
		}.Build(),
	},
	ReturnTypes: []*configv1.Field{
		configv1.Field_builder{
			Name:        "success",
			DisplayName: "Success",
			BoolField:   &configv1.BoolField{},
		}.Build(),
	},
}.Build()

var disableUserAction = v2.BatonActionSchema_builder{
	Name:        ActionDisableUser,
	DisplayName: "Disable User",
	Description: "Disables (deactivates) a Retool user account, blocking sign-in. Group memberships are kept and the account can be re-enabled.",
	ActionType:  []v2.ActionType{v2.ActionType_ACTION_TYPE_ACCOUNT_DISABLE},
	Arguments: []*configv1.Field{
		configv1.Field_builder{
			Name:        "user_id",
			DisplayName: "User",
			Description: "The Retool user to disable (deactivate).",
			IsRequired:  true,
			ResourceIdField: configv1.ResourceIdField_builder{
				Rules: configv1.ResourceIDRules_builder{
					AllowedResourceTypeIds: []string{resourceTypeUser.Id},
				}.Build(),
			}.Build(),
		}.Build(),
	},
	ReturnTypes: []*configv1.Field{
		configv1.Field_builder{
			Name:        "success",
			DisplayName: "Success",
			BoolField:   &configv1.BoolField{},
		}.Build(),
	},
}.Build()

// GlobalActions registers enable_user and disable_user as connector actions.
func (c *ConnectorImpl) GlobalActions(ctx context.Context, registry actions.ActionRegistry) error {
	if err := registry.Register(ctx, enableUserAction, c.enableUser); err != nil {
		return fmt.Errorf("failed to register %s action: %w", ActionEnableUser, err)
	}
	if err := registry.Register(ctx, disableUserAction, c.disableUser); err != nil {
		return fmt.Errorf("failed to register %s action: %w", ActionDisableUser, err)
	}
	return nil
}

func (c *ConnectorImpl) enableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	return c.setUserActive(ctx, args, true)
}

func (c *ConnectorImpl) disableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	return c.setUserActive(ctx, args, false)
}

func (c *ConnectorImpl) setUserActive(ctx context.Context, args *structpb.Struct, active bool) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if !c.client.RESTEnabled() {
		return nil, nil, status.Error(codes.Unavailable, "baton-retool: REST API is not configured; set retool-api-base-url and retool-api-token to manage accounts")
	}

	// user_id is a ResourceIdField, so the value arrives as a {resource_type_id, resource_id}
	// struct (the C1 resource picker), not a bare string.
	userResID, err := actions.RequireResourceIDArg(args, "user_id")
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "baton-retool: %v", err)
	}
	if userResID.GetResourceType() != resourceTypeUser.Id {
		return nil, nil, status.Errorf(codes.InvalidArgument, "baton-retool: user_id must reference a %q resource, got %q", resourceTypeUser.Id, userResID.GetResourceType())
	}
	rawID := userResID.GetResource()

	legacyID, err := parseObjectID(rawID)
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "baton-retool: invalid user_id %q: %v", rawID, err)
	}

	// Resolve the stable REST sid from the Postgres pool the connector already holds.
	sid, err := c.client.GetUserSID(ctx, legacyID)
	if err != nil {
		if errors.Is(err, client.ErrUserNotFound) {
			return nil, nil, status.Errorf(codes.NotFound, "baton-retool: user %q not found", rawID)
		}
		return nil, nil, status.Errorf(codes.Internal, "baton-retool: failed to resolve user %q: %v", rawID, err)
	}

	l.Debug("setting retool user state", zap.String("sid", sid), zap.Bool("active", active))

	annos, err := c.client.SetUserActive(ctx, sid, active)
	if err != nil {
		if errors.Is(err, client.ErrUserNotFound) {
			return nil, annos, status.Errorf(codes.NotFound, "baton-retool: user %q not found", rawID)
		}
		return nil, annos, status.Errorf(codes.Internal, "baton-retool: failed to set user %q active=%t: %v", rawID, active, err)
	}

	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(true),
		},
	}, annos, nil
}
