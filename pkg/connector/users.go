package connector

import (
	"context"
	"errors"
	"fmt"

	"github.com/conductorone/baton-retool/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resources "github.com/conductorone/baton-sdk/pkg/types/resource"
	_ "github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var resourceTypeUser = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

type userSyncer struct {
	resourceType      *v2.ResourceType
	client            *client.Client
	skipDisabledUsers bool
}

func (s *userSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return s.resourceType
}

func (s *userSyncer) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	pToken *pagination.Token,
) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil || parentResourceID.ResourceType != resourceTypeOrg.Id {
		return nil, "", nil, nil
	}

	orgID, err := parseObjectID(parentResourceID.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPageToken, err := s.client.ListUsersForOrg(ctx, orgID, &client.Pager{Token: pToken.Token, Size: pToken.Size}, s.skipDisabledUsers)
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range users {
		var annos annotations.Annotations

		var utStatus v2.UserTrait_Status_Status
		if o.Enabled {
			utStatus = v2.UserTrait_Status_STATUS_ENABLED
		} else {
			utStatus = v2.UserTrait_Status_STATUS_DISABLED
		}

		resourceID := formatObjectID(resourceTypeUser.Id, o.ID)
		ut, err := resources.NewUserTrait(resources.WithEmail(o.Email, true), resources.WithStatus(utStatus), resources.WithUserProfile(map[string]interface{}{
			"email":           o.Email,
			"first_name":      o.GetFirstName(),
			"last_name":       o.GetLastName(),
			"user_id":         fmt.Sprintf("%s:%s", parentResourceID.Resource, resourceID),
			"last_logged_in":  o.GetLastLoggedIn().Format("2006-01-02 15:04:05.999999999 -0700 MST"),
			"organization_id": o.OrganizationID,
			"user_name":       o.GetUserName(),
		}))
		if err != nil {
			return nil, "", nil, err
		}

		annos.Append(ut)

		ret = append(ret, &v2.Resource{
			DisplayName: fmt.Sprintf("%s %s", o.GetFirstName(), o.GetLastName()),
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     resourceID,
			},
			ParentResourceId: parentResourceID,
			Annotations:      annos,
		})
	}

	return ret, nextPageToken, nil, nil
}

func (s *userSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (s *userSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// CreateAccountCapabilityDetails advertises how accounts are provisioned. Retool sends an
// invitation on create (no connector-supplied password), so no credential is required.
func (s *userSyncer) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

// CreateAccount provisions a Retool user via the REST API. The created account is keyed by
// its Postgres legacy_id (returned in the create response), which is exactly the id the
// next full sync produces (user:<int64>) — no phantom-user reconcile needed.
func (s *userSyncer) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	credentialOptions *v2.LocalCredentialOptions,
) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	if !s.client.RESTEnabled() {
		return nil, nil, nil, status.Error(codes.Unavailable, "retool REST API is not configured; set retool-api-base-url and retool-api-token to provision accounts")
	}

	profile := accountInfo.GetProfile().AsMap()
	email := stringFromProfile(profile, "email")
	if email == "" {
		return nil, nil, nil, status.Error(codes.InvalidArgument, "email is required to create an account")
	}
	firstName := stringFromProfile(profile, "first_name")
	lastName := stringFromProfile(profile, "last_name")
	if firstName == "" || lastName == "" {
		return nil, nil, nil, status.Error(codes.InvalidArgument, "first_name and last_name are required to create an account")
	}

	user, err := s.client.CreateUser(ctx, client.CreateUserParams{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserType:  stringFromProfile(profile, "user_type"),
	})
	if err != nil {
		// Idempotent: a user with this email already exists -> return it as success.
		if errors.Is(err, client.ErrUserAlreadyExists) {
			user, err = s.client.GetUserByEmail(ctx, email)
		}
		if err != nil {
			return nil, nil, nil, status.Errorf(codes.Internal, "failed to create account for %q: %v", email, err)
		}
	}

	resource, err := s.restUserResource(user)
	if err != nil {
		return nil, nil, nil, err
	}

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: resource,
	}, nil, nil, nil
}

// Delete deprovisions a Retool user. Retool's DELETE is a soft delete (it deactivates the
// user). Idempotent: an already-deleted or unknown user is treated as success.
func (s *userSyncer) Delete(ctx context.Context, resourceId *v2.ResourceId) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if !s.client.RESTEnabled() {
		return nil, status.Error(codes.Unavailable, "retool REST API is not configured; set retool-api-base-url and retool-api-token to deprovision accounts")
	}

	legacyID, err := parseObjectID(resourceId.Resource)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id %q: %v", resourceId.Resource, err)
	}

	// Resolve the stable REST sid from the Postgres pool we already hold.
	sid, err := s.client.GetUserSID(ctx, legacyID)
	if err != nil {
		if errors.Is(err, client.ErrUserNotFound) {
			l.Debug("user already absent from Retool DB, treating delete as success", zap.Int64("legacy_id", legacyID))
			return nil, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to resolve user %d: %v", legacyID, err)
	}

	if err := s.client.DeleteUser(ctx, sid); err != nil {
		if errors.Is(err, client.ErrUserNotFound) || errors.Is(err, client.ErrUserAlreadyDisabled) {
			l.Debug("user already deprovisioned, treating delete as success", zap.String("sid", sid))
			return nil, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user %s: %v", sid, err)
	}

	return nil, nil
}

// restUserResource builds a user resource from a REST user, keyed by its Postgres
// legacy_id so it matches what a full sync produces (user:<int64>).
func (s *userSyncer) restUserResource(u *client.RESTUser) (*v2.Resource, error) {
	utStatus := v2.UserTrait_Status_STATUS_ENABLED
	if !u.Active {
		utStatus = v2.UserTrait_Status_STATUS_DISABLED
	}

	ut, err := resources.NewUserTrait(
		resources.WithEmail(u.Email, true),
		resources.WithStatus(utStatus),
		resources.WithUserProfile(map[string]interface{}{
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"user_type":  u.UserType,
		}),
	)
	if err != nil {
		return nil, err
	}

	var annos annotations.Annotations
	annos.Append(ut)

	return &v2.Resource{
		DisplayName: fmt.Sprintf("%s %s", u.FirstName, u.LastName),
		Id: &v2.ResourceId{
			ResourceType: resourceTypeUser.Id,
			Resource:     formatObjectID(resourceTypeUser.Id, u.LegacyID),
		},
		Annotations: annos,
	}, nil
}

func newUserSyncer(ctx context.Context, c *client.Client, skipDisabledUsers bool) *userSyncer {
	return &userSyncer{
		resourceType:      resourceTypeUser,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
