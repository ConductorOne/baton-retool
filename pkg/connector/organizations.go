package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	_ "github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/conductorone/baton-retool/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var resourceTypeOrg = &v2.ResourceType{
	Id:          "org",
	DisplayName: "Organization",
}

type orgSyncer struct {
	resourceType  *v2.ResourceType
	client        *client.Client
	skipPages     bool
	skipResources bool
}

func (s *orgSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return s.resourceType
}

func (s *orgSyncer) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	pToken *pagination.Token,
) ([]*v2.Resource, string, annotations.Annotations, error) {
	var annos annotations.Annotations

	orgs, nextPageToken, err := s.client.ListOrganizations(ctx, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	annos.Append(&v2.ChildResourceType{ResourceTypeId: resourceTypeUser.Id})
	annos.Append(&v2.ChildResourceType{ResourceTypeId: resourceTypeGroup.Id})
	if !s.skipPages {
		annos.Append(&v2.ChildResourceType{ResourceTypeId: resourceTypePage.Id})
	}
	if !s.skipResources {
		annos.Append(&v2.ChildResourceType{ResourceTypeId: resourceTypeResource.Id})
	}

	var ret []*v2.Resource
	for _, o := range orgs {
		ret = append(ret, &v2.Resource{
			DisplayName: o.Name,
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     formatObjectID(s.resourceType.Id, o.ID),
			},
			Annotations: annos,
		})
	}

	return ret, nextPageToken, nil, nil
}

func (s *orgSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ret []*v2.Entitlement
	var annos annotations.Annotations

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:member", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s Organization Member", resource.DisplayName),
		Description: fmt.Sprintf("Is member of the %s organization", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
		Slug:        "member",
	})

	for _, level := range accessLevels {
		ret = append(ret, &v2.Entitlement{
			Resource:    resource,
			Id:          fmt.Sprintf("entitlement:%s:universal:%s", resource.Id.Resource, level),
			DisplayName: fmt.Sprintf("%s Universal %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level])),
			Description: fmt.Sprintf("Has universal %s access on the %s organization", accessLevelDisplayNames[level], resource.DisplayName),
			GrantableTo: []*v2.ResourceType{resourceTypeUser},
			Annotations: annos,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
			Slug:        fmt.Sprintf("universal:%s", accessLevelDisplayNames[level]),
		})
		ret = append(ret, &v2.Entitlement{
			Resource:    resource,
			Id:          fmt.Sprintf("entitlement:%s:universalResource:%s", resource.Id.Resource, level),
			DisplayName: fmt.Sprintf("%s Universal Resource %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level])),
			Description: fmt.Sprintf("Has universal resource %s access on the %s organization", accessLevelDisplayNames[level], resource.DisplayName),
			GrantableTo: []*v2.ResourceType{resourceTypeUser},
			Annotations: annos,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
			Slug:        fmt.Sprintf("resource:%s", accessLevelDisplayNames[level]),
		})
		ret = append(ret, &v2.Entitlement{
			Resource:    resource,
			Id:          fmt.Sprintf("entitlement:%s:universalQueryLibrary:%s", resource.Id.Resource, level),
			DisplayName: fmt.Sprintf("%s Universal Query Library %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level])),
			Description: fmt.Sprintf("Has universal query library %s access on the %s organization", accessLevelDisplayNames[level], resource.DisplayName),
			GrantableTo: []*v2.ResourceType{resourceTypeUser},
			Annotations: annos,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
			Slug:        fmt.Sprintf("queries:%s", accessLevelDisplayNames[level]),
		})
	}

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:userList", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s User List Access", resource.DisplayName),
		Description: fmt.Sprintf("Has user list access on the %s organization", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
		Slug:        "user lists",
	})

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:auditLog", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s Audit Log Access", resource.DisplayName),
		Description: fmt.Sprintf("Has audit log access on the %s organization", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
		Slug:        "audit logs",
	})

	ret = append(ret, &v2.Entitlement{
		Resource:    resource,
		Id:          fmt.Sprintf("entitlement:%s:unpublishedRelease", resource.Id.Resource),
		DisplayName: fmt.Sprintf("%s Unpublished Release Access", resource.DisplayName),
		Description: fmt.Sprintf("Has unpublished release access on the %s organization", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{resourceTypeUser},
		Annotations: annos,
		Purpose:     v2.Entitlement_PURPOSE_VALUE_PERMISSION,
		Slug:        "unpublished releases",
	})

	return ret, "", nil, nil
}

func (s *orgSyncer) memberGrant(ctx context.Context, resource *v2.Resource, uID int64, eID string) *v2.Grant {
	entitlementID := fmt.Sprintf("entitlement:%s:%s", resource.Id.Resource, eID)
	principalID := &v2.ResourceId{
		ResourceType: resourceTypeUser.Id,
		Resource:     formatObjectID(resourceTypeUser.Id, uID),
	}
	return &v2.Grant{
		Entitlement: &v2.Entitlement{
			Id:       entitlementID,
			Resource: resource,
		},
		Principal: &v2.Resource{
			Id: principalID,
		},
		Id: fmt.Sprintf("grant:%s:%s", entitlementID, principalID.Resource),
	}
}

func (s *orgSyncer) grantsForMember(ctx context.Context, resource *v2.Resource, group *client.GroupModel, userID int64) ([]*v2.Grant, error) {
	var ret []*v2.Grant

	if group.UniversalAccess != noneLevel {
		ret = append(ret, s.memberGrant(ctx, resource, userID, fmt.Sprintf("universal:%s", group.UniversalAccess)))
	}

	if group.UniversalResourceAccess != noneLevel {
		ret = append(ret, s.memberGrant(ctx, resource, userID, fmt.Sprintf("universalResource:%s", group.UniversalResourceAccess)))
	}

	if group.UniversalQueryLibraryAccess != noneLevel {
		ret = append(ret, s.memberGrant(ctx, resource, userID, fmt.Sprintf("universalQueryLibrary:%s", group.UniversalQueryLibraryAccess)))
	}

	if group.UserListAccess {
		ret = append(ret, s.memberGrant(ctx, resource, userID, "userList"))
	}

	if group.AuditLogAccess {
		ret = append(ret, s.memberGrant(ctx, resource, userID, "auditLog"))
	}

	if group.UnpublishedReleaseAccess {
		ret = append(ret, s.memberGrant(ctx, resource, userID, "unpublishedRelease"))
	}

	ret = append(ret, s.memberGrant(ctx, resource, userID, "member"))

	return ret, nil
}

func (s *orgSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var ret []*v2.Grant

	orgID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	bag, err := parsePageToken(pToken.Token, resource.Id)
	if err != nil {
		return nil, "", nil, err
	}

	switch bag.ResourceTypeID() {
	case resourceTypeOrg.Id:
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
				ResourceID:     formatObjectID(resourceTypeGroup.Id, g.ID),
			})
		}

	case resourceTypeGroup.Id:
		gID, err := parseObjectID(bag.ResourceID())
		if err != nil {
			return nil, "", nil, err
		}

		g, err := s.client.GetGroup(ctx, gID)
		if err != nil {
			return nil, "", nil, err
		}

		members, nextPageToken, err := s.client.ListGroupMembers(ctx, g.ID, &client.Pager{Token: bag.PageToken(), Size: pToken.Size})
		if err != nil {
			return nil, "", nil, err
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, "", nil, err
		}

		for _, m := range members {
			if m.GetUserID() == 0 {
				l.Debug("member did not have user ID defined -- skipping")
				continue
			}

			memberGrants, err := s.grantsForMember(ctx, resource, g, m.GetUserID())
			if err != nil {
				return nil, "", nil, err
			}

			ret = append(ret, memberGrants...)
		}

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while processing org grants: %s", bag.ResourceTypeID())
	}

	nextPageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, nextPageToken, nil, nil
}

func newOrgSyncer(ctx context.Context, c *client.Client, skipPages bool, skipResources bool) *orgSyncer {
	return &orgSyncer{
		resourceType:  resourceTypeOrg,
		client:        c,
		skipPages:     skipPages,
		skipResources: skipResources,
	}
}
