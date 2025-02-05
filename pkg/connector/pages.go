package connector

import (
	"context"
	"errors"
	"fmt"

	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/jackc/pgx/v4"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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

	orgID, err := parseObjectID(parentResourceID.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	pages, nextPageToken, err := s.client.ListPagesForOrg(ctx, orgID, &client.Pager{Token: pToken.Token, Size: pToken.Size})
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range pages {
		ret = append(ret, &v2.Resource{
			DisplayName: o.Name,
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     formatObjectID(s.resourceType.Id, o.ID),
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
			level,
			ent.WithGrantableTo(resourceTypeGroup),
			ent.WithDisplayName(fmt.Sprintf("%s %s Access", resource.DisplayName, titleCase(accessLevelDisplayNames[level]))),
			ent.WithDescription(fmt.Sprintf("Has %s access on the %s page", accessLevelDisplayNames[level], resource.DisplayName)),
		)
		entitlement.Slug = accessLevelDisplayNames[level]

		ret = append(ret, entitlement)
	}

	return ret, "", nil, nil
}

// pageAccessLevelForGroup returns the correct access level that the group has for the page.
func (s *pageSyncer) pageAccessLevelsForGroup(ctx context.Context, page *client.PageModel, group *client.GroupModel) ([]string, error) {
	pageAccessLevels := make(map[string]struct{})
	pageAccessLevels[group.UniversalAccess] = struct{}{}

	// Check to see if a group page exists -- if it does set the access level. If not, check to see if a folder default exists and add that permission.
	// It is possible that neither of these exist.
	if groupPage, err := s.client.GetGroupPage(ctx, group.ID, page.ID); err == nil {
		pageAccessLevels[groupPage.AccessLevel] = struct{}{}
	} else if groupFolderDefault, err := s.client.GetGroupFolderDefault(ctx, group.ID, page.FolderID); err == nil {
		pageAccessLevels[groupFolderDefault.AccessLevel] = struct{}{}
	}

	if _, ok := pageAccessLevels[ownLevel]; ok {
		return []string{ownLevel, writeLevel, readLevel}, nil
	}

	if _, ok := pageAccessLevels[writeLevel]; ok {
		return []string{writeLevel, readLevel}, nil
	}

	if _, ok := pageAccessLevels[readLevel]; ok {
		return []string{readLevel}, nil
	}

	return nil, nil
}

func (s *pageSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	pageID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	page, err := s.client.GetPage(ctx, pageID)
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
	case resourceTypePage.Id:
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

		pageAccessLevels, err := s.pageAccessLevelsForGroup(ctx, page, group)
		if err != nil {
			return nil, "", nil, err
		}

		for _, level := range pageAccessLevels {
			groupId, err := rs.NewResourceID(resourceTypeGroup, formatObjectID(resourceTypeGroup.Id, group.ID))
			if err != nil {
				return nil, "", nil, err
			}

			grantExpandable := &v2.GrantExpandable{
				EntitlementIds: []string{
					fmt.Sprintf("group:%s:member", groupId.Resource),
					fmt.Sprintf("group:%s:admin", groupId.Resource),
				},
			}

			newGrant := grant.NewGrant(resource, level, groupId, grant.WithAnnotation(grantExpandable))

			ret = append(ret, newGrant)
		}

		bag.Pop()

	default:
		return nil, "", nil, fmt.Errorf("unexpected resource type while processing page grants: %s", bag.ResourceTypeID())
	}

	nextPageToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return ret, nextPageToken, nil, nil
}

func (s *pageSyncer) Grant(ctx context.Context, resource *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	if resource.Id.ResourceType != resourceTypeGroup.Id {
		return nil, nil, fmt.Errorf("unexpected resource type while processing page grant: %s", resource.Id.ResourceType)
	}

	groupID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, nil, err
	}

	pageID, err := parseObjectID(resource.Id.Resource)
	if err != nil {
		return nil, nil, err
	}

	accessLevel := entitlement.Slug

	page, err := s.client.GetGroupPage(ctx, pageID, groupID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, err
		}
	}

	// Update the group page
	if page != nil {
		if page.AccessLevel == accessLevel {
			return nil, annotations.New(&v2.GrantAlreadyExists{}), nil
		}

		err := s.client.UpdateGroupPage(ctx, page.ID, groupID, accessLevel)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Create the group page
		err := s.client.InsertGroupPage(ctx, pageID, groupID, accessLevel)
		if err != nil {
			return nil, nil, err
		}
	}

	grantExpandable := &v2.GrantExpandable{
		EntitlementIds: []string{
			fmt.Sprintf("group:%s:member", resource.Id.Resource),
			fmt.Sprintf("group:%s:admin", resource.Id.Resource),
		},
	}

	newGrant := grant.NewGrant(resource, accessLevel, resource.Id, grant.WithAnnotation(grantExpandable))

	return []*v2.Grant{newGrant}, nil, nil
}

func (s *pageSyncer) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if grant.Principal.Id.ResourceType != resourceTypeGroup.Id {
		return nil, fmt.Errorf("unexpected resource type while processing page grant: %s", grant.Principal.Id.ResourceType)
	}

	groupID, err := parseObjectID(grant.Principal.Id.Resource)
	if err != nil {
		return nil, err
	}

	pageID, err := parseObjectID(grant.Entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, err
	}

	page, err := s.client.GetGroupPage(ctx, pageID, groupID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return annotations.New(&v2.GrantAlreadyRevoked{}), nil
		} else {
			return nil, err
		}
	}

	err = s.client.DeleteGroupPage(ctx, page.ID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func newPageSyncer(ctx context.Context, c *client.Client, skipDisabledUsers bool) *pageSyncer {
	return &pageSyncer{
		resourceType:      resourceTypePage,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
