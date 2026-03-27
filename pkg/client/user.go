package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type UserModel struct {
	ID        json.Number `json:"id"`
	Email     string      `json:"email"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Active    bool        `json:"active"`
	CreatedAt string      `json:"created_at"`
	LastActive string     `json:"last_active"`
}

func (u *UserModel) GetID() int64 {
	id, _ := u.ID.Int64()
	return id
}

func (u *UserModel) GetFirstName() string {
	return u.FirstName
}

func (u *UserModel) GetLastName() string {
	return u.LastName
}

func (c *Client) ListUsers(ctx context.Context, pager *Pager, skipDisabledUsers bool) ([]*UserModel, string, error) {
	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}

	q := listQuery(offset, limit)
	data, err := c.doRequest(ctx, "GET", "/users", q, nil)
	if err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to list users: %w", err)
	}

	var resp ListResponse[*UserModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("baton-retool: failed to decode users response: %w", err)
	}

	var users []*UserModel
	for _, u := range resp.Data {
		if skipDisabledUsers && !u.Active {
			continue
		}
		users = append(users, u)
	}

	var nextPageToken string
	if resp.HasMore {
		nextPageToken = strconv.Itoa(offset + limit)
	}

	return users, nextPageToken, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*UserModel, error) {
	data, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%s", userID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to get user: %w", err)
	}

	var resp SingleResponse[*UserModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode user response: %w", err)
	}

	return resp.Data, nil
}

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Active    bool   `json:"active"`
}

func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserModel, error) {
	data, err := c.doRequest(ctx, "POST", "/users", nil, req)
	if err != nil {
		return nil, fmt.Errorf("baton-retool: failed to create user: %w", err)
	}

	var resp SingleResponse[*UserModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("baton-retool: failed to decode create user response: %w", err)
	}

	return resp.Data, nil
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	_, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/users/%s", userID), nil, nil)
	if err != nil {
		return fmt.Errorf("baton-retool: failed to delete user: %w", err)
	}

	return nil
}
