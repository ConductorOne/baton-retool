package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type OrgModel struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (c *Client) GetOrganization(ctx context.Context, orgID int64) (*OrgModel, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting organization", zap.Int64("org_id", orgID))

	var ret OrgModel
	err := pgxscan.Get(ctx, c.db, &ret, `SELECT "id", "name" FROM organizations WHERE "id"=$1`, orgID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// select id, domain, name, hostname, subdomain from organizations;.
func (c *Client) ListOrganizations(ctx context.Context, pager *Pager, organizationID *int64) ([]*OrgModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("listing organizations")

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}

	_, _ = sb.WriteString(`SELECT "id", "name" FROM organizations `)

	if organizationID != nil {
		args = append(args, *organizationID)
		_, _ = sb.WriteString(fmt.Sprintf(`WHERE "id"=$%d `, len(args)))
	}

	_, _ = sb.WriteString(`ORDER BY "id" `)

	args = append(args, limit+1)
	_, _ = sb.WriteString(fmt.Sprintf("LIMIT $%d ", len(args)))
	if offset > 0 {
		args = append(args, offset)
		_, _ = sb.WriteString(fmt.Sprintf("OFFSET $%d", len(args)))
	}

	var ret []*OrgModel
	err = pgxscan.Select(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
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
