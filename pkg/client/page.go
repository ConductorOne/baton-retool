package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type PageModel struct {
	ID             int64   `db:"id"`
	Name           string  `db:"name"`
	OrganizationID *int64  `db:"organizationId"`
	FolderID       int64   `db:"folderId"`
	PhotoUrl       *string `db:"photoUrl"`
	Description    *string `db:"description"`
}

func (g *PageModel) GetPhotoUrl() string {
	if g != nil && g.PhotoUrl != nil {
		return *g.PhotoUrl
	}

	return ""
}

func (g *PageModel) GetDescription() string {
	if g != nil && g.Description != nil {
		return *g.Description
	}

	return ""
}

func (g *PageModel) GetOrgID() int64 {
	if g != nil && g.OrganizationID != nil {
		return *g.OrganizationID
	}

	return 0
}

func (c *Client) GetPage(ctx context.Context, id int64) (*PageModel, error) {
	l := ctxzap.Extract(ctx)
	l.Info("getting page", zap.Int64("page_id", id))

	args := []interface{}{id}
	sb := &strings.Builder{}
	sb.WriteString(`select "id", "name", "organizationId", "folderId", "photoUrl", "description" from pages WHERE "deletedAt" IS NULL AND "id"=$1`)

	var ret PageModel
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) ListPagesForOrg(ctx context.Context, orgID int64, pager *Pager) ([]*PageModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Info("listing groups for org", zap.Int64("org_id", orgID))

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}
	sb.WriteString(`select "id", "name", "organizationId", "folderId", "photoUrl", "description" from pages WHERE "deletedAt" IS NULL `)

	if orgID != 0 {
		args = append(args, orgID)
		sb.WriteString(fmt.Sprintf(`AND "organizationId"=$%d `, len(args)))
	} else {
		sb.WriteString(`AND "organizationId" IS NULL `)
	}

	args = append(args, limit+1)
	sb.WriteString(fmt.Sprintf("LIMIT $%d ", len(args)))

	if offset > 0 {
		args = append(args, offset)
		sb.WriteString(fmt.Sprintf("OFFSET $%d", len(args)))
	}

	var ret []*PageModel
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
