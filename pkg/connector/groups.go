package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var resourceTypeGroup = &v2.ResourceType{
	Id:          "group",
	DisplayName: "Group",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}

const (
	adminEntitlementSlug  = "admin"
	memberEntitlementSlug = "member"
)

type groupSyncer struct {
	resourceType      *v2.ResourceType
	client            *client.Client
	skipDisabledUsers bool
}

func (s *groupSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return s.resourceType
}

func (s *groupSyncer) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	pToken *pagination.Token,
) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID != nil && parentResourceID.ResourceType != resourceTypeOrg.Id {
		return nil, "", nil, fmt.Errorf("group parent resource type must be org not %s", parentResourceID.ResourceType)
	}

	groups, nextPageToken, err := s.client.ListGroups(ctx, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource

	for _, o := range groups {
		p := make(map[string]interface{})

		options := []rs.ResourceOption{
			rs.WithGroupTrait(rs.WithGroupProfile(p)),
			rs.WithParentResourceID(parentResourceID),
		}

		resource, err := rs.NewResource(o.GetName(), s.resourceType, formatObjectID(resourceTypeGroup.Id, o.GetID()), options...)
		if err != nil {
			return nil, "", nil, err
		}

		ret = append(ret, resource)
	}

	return ret, nextPageToken, nil, nil
}

func (s *groupSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	ret := []*v2.Entitlement{
		ent.NewAssignmentEntitlement(
			resource,
			memberEntitlementSlug,
			ent.WithGrantableTo(resourceTypeUser),
			ent.WithDisplayName(fmt.Sprintf("%s Group Member", resource.DisplayName)),
			ent.WithDescription(fmt.Sprintf("Is member of the %s group", resource.DisplayName)),
		),
		ent.NewAssignmentEntitlement(
			resource,
			adminEntitlementSlug,
			ent.WithGrantableTo(resourceTypeUser),
			ent.WithDisplayName(fmt.Sprintf("%s Group Admin", resource.DisplayName)),
			ent.WithDescription(fmt.Sprintf("Is admin of the %s group", resource.DisplayName)),
		),
	}

	return ret, "", nil, nil
}

func (s *groupSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	groupID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	// Get the group with its members via the API.
	group, err := s.client.GetGroup(ctx, strconv.FormatInt(groupID, 10))
	if err != nil {
		return nil, "", nil, err
	}

	for _, m := range group.Members {
		level := "member"
		if m.IsGroupAdmin {
			level = "admin"
		}

		principalID, err := rs.NewResourceID(resourceTypeUser, formatObjectID(resourceTypeUser.Id, m.GetID()))
		if err != nil {
			return nil, "", nil, err
		}

		newGrant := grant.NewGrant(resource, level, principalID)

		ret = append(ret, newGrant)
	}

	return ret, "", nil, nil
}

func (o *groupSyncer) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"only users can be added to the group",
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)
	}

	isAdmin := entitlement.Slug == adminEntitlementSlug
	groupID, err := parseObjectID(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}

	userID := parseObjectIDString(principal.Id.Resource)

	err = o.client.AddGroupMember(ctx, strconv.FormatInt(groupID, 10), userID, isAdmin)
	if err != nil {
		if client.IsConflict(err) {
			return nil, nil
		}
		return nil, err
	}

	return nil, nil
}

func (o *groupSyncer) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	entitlement := grant.Entitlement
	principal := grant.Principal

	if principal.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"only users can be removed from the group",
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)
	}

	groupID, err := parseObjectID(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}

	userID := parseObjectIDString(principal.Id.Resource)

	err = o.client.RemoveGroupMember(ctx, strconv.FormatInt(groupID, 10), userID)
	if err != nil {
		if client.IsNotFound(err) {
			return annotations.New(&v2.GrantAlreadyRevoked{}), nil
		}
		l.Error(
			err.Error(),
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return nil, err
	}

	return nil, nil
}

func newGroupSyncer(ctx context.Context, c *client.Client, skipDisabledUsers bool) *groupSyncer {
	return &groupSyncer{
		resourceType:      resourceTypeGroup,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
