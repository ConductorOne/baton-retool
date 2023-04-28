package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-retool/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resources "github.com/conductorone/baton-sdk/pkg/types/resource"
	_ "github.com/georgysavva/scany/pgxscan"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var resourceTypeUser = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

type userSyncer struct {
	resourceType *v2.ResourceType
	client       *client.Client
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

	users, nextPageToken, err := s.client.ListUsersForOrg(ctx, orgID, &client.Pager{Token: pToken.Token, Size: pToken.Size})
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

func newUserSyncer(ctx context.Context, c *client.Client) *userSyncer {
	return &userSyncer{
		resourceType: resourceTypeUser,
		client:       c,
	}
}
