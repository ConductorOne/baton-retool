package client

import (
	"context"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type OrgModel struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// select id, domain, name, hostname, subdomain from organizations;.
func (c *Client) ListOrganizations(ctx context.Context, pager *Pager) ([]*OrgModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Info("listing organizations")

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}

	sb.WriteString(`SELECT "id", "name" FROM organizations `)

	sb.WriteString("LIMIT $1 ")
	args = append(args, limit+1)
	if offset > 0 {
		sb.WriteString("OFFSET $2")
		args = append(args, offset)
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
