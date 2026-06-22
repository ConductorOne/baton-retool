package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-retool/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resources "github.com/conductorone/baton-sdk/pkg/types/resource"

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

	users, nextPageToken, err := s.client.ListUsers(ctx, &client.Pager{Token: pToken.Token, Size: pToken.Size}, s.skipDisabledUsers)
	if err != nil {
		return nil, "", nil, err
	}

	var ret []*v2.Resource
	for _, o := range users {
		var annos annotations.Annotations

		var utStatus v2.UserTrait_Status_Status
		if o.Active {
			utStatus = v2.UserTrait_Status_STATUS_ENABLED
		} else {
			utStatus = v2.UserTrait_Status_STATUS_DISABLED
		}

		resourceID := formatObjectID(resourceTypeUser.Id, o.GetID())
		ut, err := resources.NewUserTrait(
			resources.WithEmail(o.Email, true),
			resources.WithStatus(utStatus),
			resources.WithUserProfile(map[string]interface{}{
				"email":      o.Email,
				"first_name": o.GetFirstName(),
				"last_name":  o.GetLastName(),
				"user_id":    fmt.Sprintf("%s:%s", parentResourceID.Resource, resourceID),
			}),
		)
		if err != nil {
			return nil, "", nil, err
		}

		annos.Append(ut)

		displayName := fmt.Sprintf("%s %s", o.GetFirstName(), o.GetLastName())
		if displayName == " " {
			displayName = o.Email
		}

		ret = append(ret, &v2.Resource{
			DisplayName: displayName,
			Id: &v2.ResourceId{
				ResourceType: s.resourceType.Id,
				Resource:     resourceID,
			},
			ExternalId: &v2.ExternalId{
				Id: o.ID.String(),
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

func newUserSyncer(ctx context.Context, c *client.Client, skipDisabledUsers bool) *userSyncer {
	return &userSyncer{
		resourceType:      resourceTypeUser,
		client:            c,
		skipDisabledUsers: skipDisabledUsers,
	}
}
