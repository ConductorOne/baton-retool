package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type PermissionSubject struct {
	ID   json.Number `json:"id"`
	Type string      `json:"type"` // "group" or "user"
}

type PermissionObject struct {
	ID   json.Number `json:"id"`
	Type string      `json:"type"` // "app", "folder", "resource", "resource_configuration"
}

type PermissionEntry struct {
	Object      PermissionObject `json:"object"`
	AccessLevel string           `json:"access_level"`
}

type GrantPermissionRequest struct {
	Subject     PermissionSubject `json:"subject"`
	Object      PermissionObject  `json:"object"`
	AccessLevel string            `json:"access_level"`
}

type RevokePermissionRequest struct {
	Subject PermissionSubject `json:"subject"`
	Object  PermissionObject  `json:"object"`
}

type AccessListEntry struct {
	Subject     PermissionSubject `json:"subject"`
	AccessLevel string            `json:"access_level"`
}

func (c *Client) GrantPermission(ctx context.Context, req *GrantPermissionRequest) error {
	_, err := c.doRequest(ctx, "POST", "/permissions/grant", nil, req)
	if err != nil {
		return fmt.Errorf("baton-retool: failed to grant permission: %w", err)
	}

	return nil
}

func (c *Client) RevokePermission(ctx context.Context, req *RevokePermissionRequest) error {
	_, err := c.doRequest(ctx, "POST", "/permissions/revoke", nil, req)
	if err != nil {
		return fmt.Errorf("baton-retool: failed to revoke permission: %w", err)
	}

	return nil
}

// ListObjectAccessList returns the access list for an object (app, folder, resource).
func (c *Client) ListObjectAccessList(ctx context.Context, objectType string, objectID string, pager *Pager) ([]*AccessListEntry, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	path := fmt.Sprintf("/permissions/%ss/%s", objectType, objectID)
	data, err := c.doRequest(ctx, "GET", path, q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list access list: %w", err)
	}

	var resp ListResponse[*AccessListEntry]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode access list response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}

// ListSubjectPermissions returns the list of objects with access levels that a subject has access to.
func (c *Client) ListSubjectPermissions(ctx context.Context, subjectType string, subjectID string, pager *Pager) ([]*PermissionEntry, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	path := fmt.Sprintf("/permissions/%ss/%s/objects", subjectType, subjectID)
	data, err := c.doRequest(ctx, "GET", path, q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list subject permissions: %w", err)
	}

	var resp ListResponse[*PermissionEntry]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode subject permissions response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}
