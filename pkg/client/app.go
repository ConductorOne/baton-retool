package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// AppModel represents a Retool app (formerly "page" in the DB schema).
type AppModel struct {
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	FolderID    json.Number `json:"folder_id"`
	Description string      `json:"description"`
}

func (a *AppModel) GetID() int64 {
	id, _ := a.ID.Int64()
	return id
}

func (c *Client) ListApps(ctx context.Context, pager *Pager) ([]*AppModel, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	data, err := c.doRequest(ctx, "GET", "/apps", q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list apps: %w", err)
	}

	var resp ListResponse[*AppModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode apps response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}

func (c *Client) GetApp(ctx context.Context, appID string) (*AppModel, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/apps/%s", appID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to get app: %w", err)
	}

	var resp SingleResponse[*AppModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode app response: %w", err)
	}

	return resp.Data, nil
}
