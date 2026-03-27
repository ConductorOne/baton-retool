package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

var resourceTypePage = &v2.ResourceType{
	Id:          "page",
	DisplayName: "Page",
}

type pageSyncer struct {
	resourceType      *v2.ResourceType
	client            *client.Client
	skipDisabledUsers bool
}

func (s *pageSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return s.resourceType
}

func (s *pageSyncer) List(
	ctx context.Context,
	parentResourceID *v2.ResourceId,
	pToken *pagination.Token,
) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil || parentResourceID.ResourceType != resourceTypeOrg.Id {
		return nil, "", nil, nil
	}

	apps, nextPageToken, err := s.client.ListApps(ctx, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range apps {
		ret = append(ret, &v2.Resource{
			DisplayName: o.Name,
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     formatObjectID(s.resourceType.Id, o.GetID()),
			},
			ParentResourceId: parentResourceID,
		})
	}

	return ret, nextPageToken, nil, nil
}

func (s *pageSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ret []*v2.Entitlement

	for _, level := range accessLevels {
		entitlement := ent.NewPermissionEntitlement(
			resource,
			fmt.Sprintf("%s:%s", "group", level),
			ent.WithGrantableTo(resourceTypeGroup),
			ent.WithDisplayName(fmt.Sprintf("%s %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level]))),
			ent.WithDescription(fmt.Sprintf("Has %s access on the %s page", accessLevelDisplayNames[level], resource.DisplayName)),
			ent.WithAnnotation(&v2.EntitlementImmutable{}),
		)

		ret = append(ret, entitlement)
	}

	for _, level := range accessLevels {
		entitlement := ent.NewPermissionEntitlement(
			resource,
			fmt.Sprintf("%s:%s", "user", level),
			ent.WithGrantableTo(resourceTypeUser),
			ent.WithDisplayName(fmt.Sprintf("User can %s on %s", titleCase(accessLevelDisplayNames[level]), resource.DisplayName)),
			ent.WithDescription(fmt.Sprintf("Has %s access on the %s page", accessLevelDisplayNames[level], resource.DisplayName)),
		)

		ret = append(ret, entitlement)
	}

	return ret, "", nil, nil
}

func (s *pageSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	appID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	appIDStr := strconv.FormatInt(appID, 10)

	// Use the permissions API to get the access list for this app.
	entries, nextPageToken, err := s.client.ListObjectAccessList(ctx, "app", appIDStr, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	for _, entry := range entries {
		accessLevel := entry.AccessLevel
		if accessLevel == noneLevel || accessLevel == "" {
			continue
		}

		subjectType := entry.Subject.Type
		subjectID := entry.Subject.ID

		switch subjectType {
		case "group":
			groupID, err := subjectID.Int64()
			if err != nil {
				continue
			}

			principalID, err := rs.NewResourceID(resourceTypeGroup, formatObjectID(resourceTypeGroup.Id, groupID))
			if err != nil {
				return nil, "", nil, err
			}

			grantExpandable := &v2.GrantExpandable{
				EntitlementIds: []string{
					fmt.Sprintf("group:%s:member", principalID.Resource),
					fmt.Sprintf("group:%s:admin", principalID.Resource),
				},
			}

			newGroupGrant := grant.NewGrant(resource, fmt.Sprintf("group:%s", accessLevel), principalID, grant.WithAnnotation(grantExpandable))
			ret = append(ret, newGroupGrant)

		case "user":
			userID, err := subjectID.Int64()
			if err != nil {
				continue
			}

			principalID, err := rs.NewResourceID(resourceTypeUser, formatObjectID(resourceTypeUser.Id, userID))
			if err != nil {
				return nil, "", nil, err
			}

			newUserGrant := grant.NewGrant(resource, fmt.Sprintf("user:%s", accessLevel), principalID)
			ret = append(ret, newUserGrant)
		}
	}

	return ret, nextPageToken, nil, nil
}

func (s *pageSyncer) Grant(ctx context.Context, resource *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	appID, err := parseObjectID(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, nil, err
	}

	splitV := strings.Split(entitlement.Id, ":")
	if len(splitV) != 4 {
		return nil, nil, fmt.Errorf("unexpected entitlement ID format while processing page grant: %s", entitlement.Id)
	}
	accessLevel := splitV[len(splitV)-1]

	switch resource.Id.ResourceType {
	case resourceTypeUser.Id:
		userID, err := parseObjectID(resource.Id.Resource)
		if err != nil {
			return nil, nil, err
		}

		err = s.client.GrantPermission(ctx, &client.GrantPermissionRequest{
			Subject: client.PermissionSubject{
				ID:   json.Number(strconv.FormatInt(userID, 10)),
				Type: "user",
			},
			Object: client.PermissionObject{
				ID:   json.Number(strconv.FormatInt(appID, 10)),
				Type: "app",
			},
			AccessLevel: accessLevel,
		})
		if err != nil {
			if client.IsConflict(err) {
				return nil, annotations.New(&v2.GrantAlreadyExists{}), nil
			}
			return nil, nil, err
		}

		newGrant := grant.NewGrant(entitlement.Resource, fmt.Sprintf("user:%s", accessLevel), resource.Id)
		return []*v2.Grant{newGrant}, nil, nil

	case resourceTypeGroup.Id:
		groupID, err := parseObjectID(resource.Id.Resource)
		if err != nil {
			return nil, nil, err
		}

		err = s.client.GrantPermission(ctx, &client.GrantPermissionRequest{
			Subject: client.PermissionSubject{
				ID:   json.Number(strconv.FormatInt(groupID, 10)),
				Type: "group",
			},
			Object: client.PermissionObject{
				ID:   json.Number(strconv.FormatInt(appID, 10)),
				Type: "app",
			},
			AccessLevel: accessLevel,
		})
		if err != nil {
			if client.IsConflict(err) {
				return nil, annotations.New(&v2.GrantAlreadyExists{}), nil
			}
			return nil, nil, err
		}

		grantExpandable := &v2.GrantExpandable{
			EntitlementIds: []string{
				fmt.Sprintf("group:%s:member", resource.Id.Resource),
				fmt.Sprintf("group:%s:admin", resource.Id.Resource),
			},
		}

		newGrant := grant.NewGrant(entitlement.Resource, fmt.Sprintf("group:%s", accessLevel), resource.Id, grant.WithAnnotation(grantExpandable))
		return []*v2.Grant{newGrant}, nil, nil

	default:
		return nil, nil, fmt.Errorf("unexpected resource type while processing page grant: %s", resource.Id.ResourceType)
	}
}

func (s *pageSyncer) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	appID, err := parseObjectID(grant.Entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}

	principal := grant.Principal

	switch principal.Id.ResourceType {
	case resourceTypeUser.Id:
		userID, err := parseObjectID(principal.Id.Resource)
		if err != nil {
			return nil, err
		}

		err = s.client.RevokePermission(ctx, &client.RevokePermissionRequest{
			Subject: client.PermissionSubject{
				ID:   json.Number(strconv.FormatInt(userID, 10)),
				Type: "user",
			},
			Object: client.PermissionObject{
				ID:   json.Number(strconv.FormatInt(appID, 10)),
				Type: "app",
			},
		})
		if err != nil {
			if client.IsNotFound(err) {
				return annotations.New(&v2.GrantAlreadyRevoked{}), nil
			}
			return nil, err
		}

		return nil, nil

	case resourceTypeGroup.Id:
		groupID, err := parseObjectID(principal.Id.Resource)
		if err != nil {
			return nil, err
		}

		err = s.client.RevokePermission(ctx, &client.RevokePermissionRequest{
			Subject: client.PermissionSubject{
				ID:   json.Number(strconv.FormatInt(groupID, 10)),
				Type: "group",
			},
			Object: client.PermissionObject{
				ID:   json.Number(strconv.FormatInt(appID, 10)),
				Type: "app",
			},
		})
		if err != nil {
			if client.IsNotFound(err) {
				return annotations.New(&v2.GrantAlreadyRevoked{}), nil
			}
			return nil, err
		}

		return nil, nil

	default:
		return nil, fmt.Errorf("unexpected resource type while processing page revoke: %s", principal.Id.ResourceType)
	}
}

func newPageSyncer(c *client.Client, skipDisabledUsers bool) *pageSyncer {
	return &pageSyncer{
		resourceType:      resourceTypePage,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
