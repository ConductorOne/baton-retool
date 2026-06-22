package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type FolderModel struct {
	ID             json.Number `json:"id"`
	Name           string      `json:"name"`
	ParentFolderID json.Number `json:"parent_folder_id"`
	FolderType     string      `json:"folder_type"` // "app" or "resource"
	IsSystemFolder bool        `json:"is_system_folder"`
}

func (f *FolderModel) GetID() int64 {
	id, _ := f.ID.Int64()
	return id
}

func (c *Client) ListFolders(ctx context.Context, folderType string, pager *Pager) ([]*FolderModel, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	if folderType != "" {
		q.Set("folder_type", folderType)
	}

	data, err := c.doRequest(ctx, "GET", "/folders", q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list folders: %w", err)
	}

	var resp ListResponse[*FolderModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode folders response: %w", err)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return resp.Data, nextPageToken, nil
}
