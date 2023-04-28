package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resources "github.com/conductorone/baton-sdk/pkg/types/resource"
)

var resourceTypeGroup = &v2.ResourceType{
	Id:          "group",
	DisplayName: "Group",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}

type groupSyncer struct {
	resourceType *v2.ResourceType
	client       *client.Client
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
		var annos annotations.Annotations

		p := make(map[string]interface{})

		if o.OrganizationID != nil {
			p["organization_id"] = o.GetOrgID()
		}

		gt, err := resources.NewGroupTrait(resources.WithGroupProfile(p))
		if err != nil {
			return nil, "", nil, err
		}

		annos.Append(gt)

		ret = append(ret, &v2.Resource{
			DisplayName: o.GetName(),
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     formatObjectID(s.resourceType.Id, o.ID),
			},
			ParentResourceId: parentResourceID,
			Annotations:      annos,
		})
	}

	return ret, nextPageToken, nil, nil
}

func (s *groupSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ret []*v2.Entitlement
	var annos annotations.Annotations

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:member", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s Group Member", resource.DisplayName),
		Description: fmt.Sprintf("Is member of the %s organization", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
		Slug:        "member",
	})

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:admin", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s Group Admin", resource.DisplayName),
		Description: fmt.Sprintf("Is admin of the %s group", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
		Slug:        "admin",
	})
	return ret, "", nil, nil
}

func (s *groupSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	groupID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	members, nextPageToken, err := s.client.ListGroupMembers(ctx, groupID, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	for _, m := range members {
		level := "member"
		if m.IsAdmin {
			level = "admin"
		}
		entitlementID := fmt.Sprintf("entitlement:%s:%s", resource.Id.Resource, level)
		principalID := formatObjectID(resourceTypeUser.Id, m.GetUserID())

		ret = append(ret, &v2.Grant{
			Entitlement: &v2.Entitlement{
				Id:       entitlementID,
				Resource: resource,
			},
			Principal: &v2.Resource{
				Id: &v2.ResourceId{
					ResourceType: resourceTypeUser.Id,
					Resource:     principalID,
				},
			},
			Id: fmt.Sprintf("grant:%s:%s", entitlementID, principalID),
		})
	}

	return ret, nextPageToken, nil, nil
}

func newGroupSyncer(ctx context.Context, c *client.Client) *groupSyncer {
	return &groupSyncer{
		resourceType: resourceTypeGroup,
		client:       c,
	}
}
