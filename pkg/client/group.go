package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type GroupModel struct {
	ID                          json.Number    `json:"id"`
	Name                        string         `json:"name"`
	UniversalAppAccess          string         `json:"universal_app_access"`
	UniversalResourceAccess     string         `json:"universal_resource_access"`
	UniversalQueryLibraryAccess string         `json:"universal_query_library_access"`
	UserListAccess              bool           `json:"user_list_access"`
	AuditLogAccess              bool           `json:"audit_log_access"`
	UnpublishedReleaseAccess    bool           `json:"unpublished_release_access"`
	Members                     []*GroupMember `json:"members"`
}

func (g *GroupModel) GetID() int64 {
	id, _ := g.ID.Int64()
	return id
}

func (g *GroupModel) GetName() string {
	return g.Name
}

type GroupMember struct {
	ID           json.Number `json:"id"`
	IsGroupAdmin bool        `json:"is_group_admin"`
}

func (m *GroupMember) GetID() int64 {
	id, _ := m.ID.Int64()
	return id
}

func (c *Client) ListGroups(ctx context.Context, pager *Pager) ([]*GroupModel, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	data, err := c.doRequest(ctx, "GET", "/groups", q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list groups: %w", err)
	}

	var resp ListResponse[*GroupModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode groups response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}

func (c *Client) GetGroup(ctx context.Context, groupID string) (*GroupModel, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/groups/%s", groupID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to get group: %w", err)
	}

	var resp SingleResponse[*GroupModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode group response: %w", err)
	}

	return resp.Data, nil
}

type AddGroupMembersRequest struct {
	Members []AddGroupMember `json:"members"`
}

type AddGroupMember struct {
	ID           json.Number `json:"id"`
	IsGroupAdmin bool        `json:"is_group_admin"`
}

func (c *Client) AddGroupMember(ctx context.Context, groupID string, userID string, isAdmin bool) error {
	req := &AddGroupMembersRequest{
		Members: []AddGroupMember{
			{
				ID:           json.Number(userID),
				IsGroupAdmin: isAdmin,
			},
		},
	}

	_, err := c.doRequest(ctx, "POST", fmt.Sprintf("/groups/%s/members", groupID), nil, req)
	if err != nil {
		return fmt.Errorf("baton-retool: failed to add group member: %w", err)
	}

	return nil
}

func (c *Client) RemoveGroupMember(ctx context.Context, groupID string, userID string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/groups/%s/members/%s", groupID, userID), nil, nil)
	if err != nil {
		return fmt.Errorf("baton-retool: failed to remove group member: %w", err)
	}

	return nil
}
