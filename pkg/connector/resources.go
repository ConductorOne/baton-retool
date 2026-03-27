package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
)

var resourceTypeResource = &v2.ResourceType{
	Id:          "resource",
	DisplayName: "Resource",
}

type resourceSyncer struct {
	resourceType *v2.ResourceType
	client       *client.Client
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

	resources, nextPageToken, err := s.client.ListResources(ctx, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range resources {
		displayName := o.GetDisplayName()
		if displayName == "" {
			displayName = o.Name
		}
		ret = append(ret, &v2.Resource{
			DisplayName: fmt.Sprintf("%s (%s)", displayName, o.Type),
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     formatObjectID(s.resourceType.Id, o.GetID()),
			},
			ParentResourceId: parentResourceID,
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
			GrantableTo: []*v2.ResourceType{resourceTypeGroup},
			Annotations: annos,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
			Slug:        accessLevelDisplayNames[level],
		})
	}

	return ret, "", nil, nil
}

func (s *resourceSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	resourceID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	resourceIDStr := strconv.FormatInt(resourceID, 10)

	entries, nextPageToken, err := s.client.ListObjectAccessList(ctx, "resource", resourceIDStr, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	for _, entry := range entries {
		accessLevel := entry.AccessLevel
		if accessLevel == noneLevel || accessLevel == "" {
			continue
		}

		if accessLevel != readLevel && accessLevel != writeLevel {
			continue
		}

		subjectType := entry.Subject.Type
		subjectID := entry.Subject.ID

		if subjectType == "group" {
			gID, err := subjectID.Int64()
			if err != nil {
				continue
			}

			principalID, err := rs.NewResourceID(resourceTypeGroup, formatObjectID(resourceTypeGroup.Id, gID))
			if err != nil {
				return nil, "", nil, err
			}

			levels := []string{accessLevel}
			if accessLevel == writeLevel {
				levels = append(levels, readLevel)
			}

			for _, level := range levels {
				eID := fmt.Sprintf("entitlement:%s:%s", resource.Id.Resource, level)
				newGrant := grant.NewGrant(resource, level, principalID)
				newGrant.Id = fmt.Sprintf("grant:%s:%s", eID, principalID.Resource)

				ret = append(ret, newGrant)
			}
		}
	}

	return ret, nextPageToken, nil, nil
}

func newResourceSyncer(ctx context.Context, c *client.Client) *resourceSyncer {
	return &resourceSyncer{
		resourceType: resourceTypeResource,
		client:       c,
	}
}
