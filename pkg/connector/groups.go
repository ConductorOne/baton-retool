package connector

import (
	"context"
	"fmt"

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
	var orgID int64
	var err error

	if parentResourceID != nil {
		if parentResourceID.ResourceType != resourceTypeOrg.Id {
			return nil, "", nil, fmt.Errorf("group parent resource type must be org not %s", parentResourceID.ResourceType)
		}
		orgID, err = parseObjectID(parentResourceID.Resource)
		if err != nil {
			return nil, "", nil, err
		}
	}

	groups, nextPageToken, err := s.client.ListGroupsForOrg(ctx, orgID, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource

	for _, o := range groups {
		p := make(map[string]interface{})

		if o.OrganizationID != nil {
			p["organization_id"] = o.GetOrgID()
		}

		options := []rs.ResourceOption{
			rs.WithGroupTrait(rs.WithGroupProfile(p)),
			rs.WithParentResourceID(parentResourceID),
		}

		resource, err := rs.NewResource(o.GetName(), s.resourceType, formatObjectID(resourceTypeGroup.Id, o.ID), options...)
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

	members, nextPageToken, err := s.client.ListGroupMembers(ctx, groupID, &client.Pager{Token: pToken.Token, Size: pToken.Size}, s.skipDisabledUsers)
	if err != nil {
		return nil, "", nil, err
	}

	for _, m := range members {
		level := "member"
		if m.IsAdmin {
			level = "admin"
		}

		principalID, err := rs.NewResourceID(resourceTypeUser, formatObjectID(resourceTypeUser.Id, m.GetUserID()))
		if err != nil {
			return nil, "", nil, err
		}

		newGrant := grant.NewGrant(resource, level, principalID)

		ret = append(ret, newGrant)
	}

	return ret, nextPageToken, nil, nil
}

func (o *groupSyncer) Grant(ctx context.Context, principial *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principial.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"only users can be added to the group",
			zap.String("principal_id", principial.Id.Resource),
			zap.String("principal_type", principial.Id.ResourceType),
		)
	}

	isAdminNewValue := entitlement.Slug == adminEntitlementSlug
	groupID, err := parseObjectID(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}
	userID, err := parseObjectID(principial.Id.Resource)
	if err != nil {
		return nil, err
	}

	member, err := o.client.GetGroupMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	if member == nil {
		err = o.client.AddGroupMember(ctx, groupID, userID, isAdminNewValue)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	if member.IsAdmin == isAdminNewValue {
		return nil, nil
	}

	_, err = o.client.UpdateGroupMember(ctx, groupID, userID, isAdminNewValue)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (o *groupSyncer) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	entitlement := grant.Entitlement
	principial := grant.Principal

	if principial.Id.ResourceType != resourceTypeUser.Id {
		l.Warn(
			"only users can be added to the group",
			zap.String("principal_id", principial.Id.Resource),
			zap.String("principal_type", principial.Id.ResourceType),
		)
	}

	groupID, err := parseObjectID(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}
	userID, err := parseObjectID(principial.Id.Resource)
	if err != nil {
		return nil, err
	}

	err = o.client.RemoveGroupMember(ctx, groupID, userID)
	if err != nil {
		l.Error(
			err.Error(),
			zap.String("principal_id", principial.Id.Resource),
			zap.String("principal_type", principial.Id.ResourceType),
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
