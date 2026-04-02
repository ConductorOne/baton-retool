package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type ResourceModel struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	DisplayName string      `json:"display_name"`
}

func (r *ResourceModel) GetID() int64 {
	id, _ := r.ID.Int64()
	return id
}

func (r *ResourceModel) GetDisplayName() string {
	if r.DisplayName != "" {
		return r.DisplayName
	}
	return r.Name
}

func (c *Client) ListResources(ctx context.Context, pager *Pager) ([]*ResourceModel, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	data, err := c.doRequest(ctx, "GET", "/resources", q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list resources: %w", err)
	}

	var resp ListResponse[*ResourceModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode resources response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}

func (c *Client) GetResource(ctx context.Context, resourceID string) (*ResourceModel, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/resources/%s", resourceID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to get resource: %w", err)
	}

	var resp SingleResponse[*ResourceModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode resource response: %w", err)
	}

	return resp.Data, nil
}
