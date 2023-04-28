package client

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type ResourceModel struct {
	ID               int64   `db:"id"`
	OrganizationID   int64   `db:"organizationId"`
	Name             string  `db:"name"`
	Type             string  `db:"type"`
	DisplayName      *string `db:"displayName"`
	EnvironmentID    *string `db:"environmentId"`
	ResourceFolderID *int64  `db:"resourceFolderId"`
}

func (u *ResourceModel) GetDisplayName() string {
	if u != nil && u.DisplayName != nil {
		return *u.DisplayName
	}

	return ""
}

func (u *ResourceModel) GetEnvironmentID() string {
	if u != nil && u.EnvironmentID != nil {
		return *u.EnvironmentID
	}

	return ""
}

func (g *ResourceModel) GetResourceFolderID() int64 {
	if g != nil && g.ResourceFolderID != nil {
		return *g.ResourceFolderID
	}

	return 0
}

func (c *Client) GetResource(ctx context.Context, id int64) (*ResourceModel, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting resource", zap.Int64("resource_id", id))

	args := []interface{}{id}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "name", "type", "displayName", "environmentId", "resourceFolderId" from resources WHERE "id"=$1 `)

	var ret ResourceModel
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) ListResourcesForOrg(ctx context.Context, orgID int64, pager *Pager) ([]*ResourceModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("listing resources for org", zap.Int64("org_id", orgID))

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "name", "type", "displayName", "environmentId", "resourceFolderId" from resources WHERE "organizationId"=$1 `)
	args = append(args, orgID)
	_, _ = sb.WriteString("LIMIT $2 ")
	args = append(args, limit+1)
	if offset > 0 {
		_, _ = sb.WriteString("OFFSET $3")
		args = append(args, offset)
	}

	var ret []*ResourceModel
	err = pgxscan.Select(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}

	var nextPageToken string
	if len(ret) > limit {
		offset += limit
		nextPageToken = strconv.Itoa(offset)
		ret = ret[:limit]
	}

	return ret, nextPageToken, nil
}
