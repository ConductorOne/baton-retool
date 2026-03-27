package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type OrgModel struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
}

func (o *OrgModel) GetID() int64 {
	id, _ := o.ID.Int64()
	return id
}

func (c *Client) GetOrganization(ctx context.Context) (*OrgModel, error) {
	data, err := c.doRequest(ctx, "GET", "/organization", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to get organization: %w", err)
	}

	var resp SingleResponse[*OrgModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode organization response: %w", err)
	}

	return resp.Data, nil
}
