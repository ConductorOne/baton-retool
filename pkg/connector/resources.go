package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-retool/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	_ "github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var resourceTypeResource = &v2.ResourceType{
	Id:          "resource",
	DisplayName: "Resource",
}

type resourceSyncer struct {
	resourceType      *v2.ResourceType
	client            *client.Client
	skipDisabledUsers bool
}

func (s *resourceSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return s.resourceType
}

func (s *resourceSyncer) List(
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

	resources, nextPageToken, err := s.client.ListResourcesForOrg(ctx, orgID, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range resources {
		var annos annotations.Annotations
		displayName := o.GetDisplayName()
		if displayName == "" {
			displayName = o.Name
		}
		ret = append(ret, &v2.Resource{
			DisplayName: fmt.Sprintf("%s (%s)", displayName, o.Type),
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

func (s *resourceSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ret []*v2.Entitlement
	var annos annotations.Annotations

	for _, level := range resourceAccessLevels {
		ret = append(ret, &v2.Entitlement{
			Resource:    resource,
			Id:          fmt.Sprintf("entitlement:%s:%s", resource.Id.Resource, level),
			DisplayName: fmt.Sprintf("%s %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level])),
			Description: fmt.Sprintf("Has %s access on the %s resource", accessLevelDisplayNames[level], resource.DisplayName),
			GrantableTo: []*v2.ResourceType{resourceTypeUser},
			Annotations: annos,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
			Slug:        accessLevelDisplayNames[level],
		})
	}

	return ret, "", nil, nil
}

// pageAccessLevelForGroup returns the correct access level that the group has for the page.
func (s *resourceSyncer) resourceAccessLevelsForGroup(ctx context.Context, rsrc *client.ResourceModel, group *client.GroupModel) ([]string, error) {
	retAccessLevels := make(map[string]struct{})
	retAccessLevels[group.UniversalResourceAccess] = struct{}{}

	// Check to see if a group rsrc exists -- if it does set the access level. If not, check to see if a folder default exists and add that permission.
	// It is possible that neither of these exist.
	if groupPage, err := s.client.GetGroupResource(ctx, group.ID, rsrc.ID); err == nil {
		retAccessLevels[groupPage.AccessLevel] = struct{}{}
	} else if groupFolderDefault, err := s.client.GetGroupResourceFolderDefault(ctx, group.ID, rsrc.GetResourceFolderID()); err == nil {
		retAccessLevels[groupFolderDefault.AccessLevel] = struct{}{}
	}

	if _, ok := retAccessLevels[writeLevel]; ok {
		return []string{writeLevel, readLevel}, nil
	}

	if _, ok := retAccessLevels[readLevel]; ok {
		return []string{readLevel}, nil
	}

	return nil, nil
}

func (s *resourceSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var ret []*v2.Grant

	objectID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	obj, err := s.client.GetResource(ctx, objectID)
	if err != nil {
		return nil, "", nil, err
	}

	orgID, err := parseObjectID(resource.ParentResourceId.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	bag, err := parsePageToken(pToken.Token, resource.Id)
	if err != nil {
		return nil, "", nil, err
	}

	switch bag.ResourceTypeID() {
	case resourceTypeResource.Id:
		groups, nextPageToken, err := s.client.ListGroupsForOrg(ctx, orgID, &client.Pager{Token: bag.PageToken(), Size: pToken.Size})
		if err != nil {
			return nil, "", nil, err
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}

		// push pagination state for each group
		for _, g := range groups {
			bag.Push(pagination.PageState{
				ResourceTypeID: resourceTypeGroup.Id,
				ResourceID:     formatGroupObjectID(g.ID),
			})
		}

	case resourceTypeGroup.Id:
		groupID, err := parseGroupObjectID(bag.ResourceID())
		if err != nil {
			return nil, "", nil, err
		}

		group, err := s.client.GetGroup(ctx, groupID)
		if err != nil {
			return nil, "", nil, err
		}

		members, nextPageToken, err := s.client.ListGroupMembers(ctx, group.ID, &client.Pager{Token: bag.PageToken(), Size: pToken.Size}, s.skipDisabledUsers)
		if err != nil {
			return nil, "", nil, err
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}

		levels, err := s.resourceAccessLevelsForGroup(ctx, obj, group)
		if err != nil {
			return nil, "", nil, err
		}

		for _, m := range members {
			if m.GetUserID() == 0 {
				l.Debug("member did not have user ID defined -- skipping")
				continue
			}

			principalID := &v2.ResourceId{
				ResourceType: resourceTypeUser.Id,
				Resource:     formatObjectID(resourceTypeUser.Id, m.GetUserID()),
			}

			for _, level := range levels {
				eID := fmt.Sprintf("entitlement:%s:%s", resource.Id.Resource, level)
				ret = append(ret, &v2.Grant{
					Entitlement: &v2.Entitlement{
						Id:       eID,
						Resource: resource,
					},
					Principal: &v2.Resource{
						Id: principalID,
					},
					Id: fmt.Sprintf("grant:%s:%s", eID, principalID.Resource),
				})
			}
		}

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while processing resource grants: %s", bag.ResourceTypeID())
	}

	nextPageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, nextPageToken, nil, nil
}

func newResourceSyncer(ctx context.Context, c *client.Client, skipDisabledUsers bool) *resourceSyncer {
	return &resourceSyncer{
		resourceType:      resourceTypeResource,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
