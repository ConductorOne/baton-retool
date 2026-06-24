package client

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type UserModel struct {
	ID              int64      `db:"id"`
	Email           string     `db:"email"`
	FirstName       *string    `db:"firstName"`
	LastName        *string    `db:"lastName"`
	ProfilePhotoURL *string    `db:"profilePhotoUrl"`
	UserName        *string    `db:"userName"`
	Enabled         bool       `db:"enabled"`
	LastLoggedIn    *time.Time `db:"lastLoggedIn"`
	OrganizationID  int64      `db:"organizationId"`
}

func (u *UserModel) GetFirstName() string {
	if u != nil && u.FirstName != nil {
		return *u.FirstName
	}

	return ""
}

func (u *UserModel) GetLastName() string {
	if u != nil && u.LastName != nil {
		return *u.LastName
	}

	return ""
}

func (u *UserModel) GetProfilePhotoUrl() string {
	if u != nil && u.ProfilePhotoURL != nil {
		return *u.ProfilePhotoURL
	}

	return ""
}

func (u *UserModel) GetUserName() string {
	if u != nil && u.UserName != nil {
		return *u.UserName
	}

	return ""
}

func (u *UserModel) GetLastLoggedIn() time.Time {
	if u != nil && u.LastLoggedIn != nil {
		return *u.LastLoggedIn
	}

	return time.Time{}
}

// ErrUserNotFound is returned by GetUserSID when no user row matches the legacy id.
var ErrUserNotFound = errors.New("user not found")

// GetUserSID resolves a Postgres user id (the int64 that baton keys `user:<int64>` on,
// surfaced as `legacy_id` over REST) to its stable `sid` (`user_<uuid>`), which is the id
// every REST user endpoint addresses users by. Because the connector always holds the
// Postgres pool, this is the robust int64->UUID join — no mutable-email lookup required.
func (c *Client) GetUserSID(ctx context.Context, legacyID int64) (string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("resolving user sid", zap.Int64("legacy_id", legacyID))

	var sid string
	err := pgxscan.Get(ctx, c.db, &sid, `SELECT "sid" FROM users WHERE "id"=$1`, legacyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || pgxscan.NotFound(err) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return sid, nil
}

func (c *Client) ListUsersForOrg(ctx context.Context, orgID int64, pager *Pager, skipDisabledUsers bool) ([]*UserModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("listing users for org", zap.Int64("org_id", orgID))

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}
	_, _ = sb.WriteString(`SELECT "id", "email", "firstName", "lastName", "profilePhotoUrl", "enabled", "userName", "organizationId", "lastLoggedIn" from users WHERE "organizationId"=$1 `)
	args = append(args, orgID)
	if skipDisabledUsers {
		_, _ = sb.WriteString("AND enabled = true ")
	}
	_, _ = sb.WriteString(`ORDER BY "id" `)
	_, _ = sb.WriteString("LIMIT $2 ")
	args = append(args, limit+1)
	if offset > 0 {
		_, _ = sb.WriteString("OFFSET $3")
		args = append(args, offset)
	}

	var ret []*UserModel
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
