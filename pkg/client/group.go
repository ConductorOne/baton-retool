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

type GroupModel struct {
	ID                          int64   `db:"id"`
	Name                        *string `db:"name"`
	OrganizationID              *int64  `db:"organizationId"`
	UniversalAccess             string  `db:"universalAccess"`
	UniversalResourceAccess     string  `db:"universalResourceAccess"`
	UniversalQueryLibraryAccess string  `db:"universalQueryLibraryAccess"`
	UserListAccess              bool    `db:"userListAccess"`
	AuditLogAccess              bool    `db:"auditLogAccess"`
	UnpublishedReleaseAccess    bool    `db:"unpublishedReleaseAccess"`
}

func (g *GroupModel) GetName() string {
	if g != nil && g.Name != nil {
		return *g.Name
	}

	return ""
}

func (g *GroupModel) GetOrgID() int64 {
	if g != nil && g.OrganizationID != nil {
		return *g.OrganizationID
	}

	return 0
}

type GroupPage struct {
	ID          int64  `db:"id"`
	AccessLevel string `db:"accessLevel"`
}

type GroupFolderDefault struct {
	ID          int64  `db:"id"`
	AccessLevel string `db:"accessLevel"`
}

type GroupResource struct {
	ID          int64  `db:"id"`
	AccessLevel string `db:"accessLevel"`
}

type GroupResourceFolderDefault struct {
	ID          int64  `db:"id"`
	AccessLevel string `db:"accessLevel"`
}

type GroupMember struct {
	Id      int64  `db:"id"`
	UserID  *int64 `db:"userId"`
	GroupID *int64 `db:"groupId"`
	IsAdmin bool   `db:"isAdmin"`
	Enabled bool   `db:"enabled"`
}

func (g *GroupMember) GetUserID() int64 {
	if g != nil && g.UserID != nil {
		return *g.UserID
	}

	return 0
}

func (g *GroupMember) GetGroupID() int64 {
	if g != nil && g.GroupID != nil {
		return *g.GroupID
	}

	return 0
}

func (c *Client) GetGroupPage(ctx context.Context, groupID int64, pageID int64) (*GroupPage, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group page", zap.Int64("group_id", groupID), zap.Int64("page_id", pageID))

	args := []interface{}{groupID, pageID}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "accessLevel" from group_pages WHERE "groupId"=$1 AND "pageId"=$2`)

	var ret GroupPage
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) InsertGroupPage(ctx context.Context, groupID int64, pageID int64, accessLevel string) error {
	l := ctxzap.Extract(ctx)
	l.Debug("inserting group page", zap.Int64("group_id", groupID), zap.Int64("page_id", pageID), zap.String("access_level", accessLevel))

	args := []interface{}{groupID, pageID, accessLevel}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`INSERT INTO group_pages ("groupId", "pageId", "accessLevel", "createdAt", "updatedAt") VALUES ($1, $2, $3, NOW(), NOW())`)

	if _, err := c.db.Exec(ctx, sb.String(), args...); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateGroupPage(ctx context.Context, id int64, accessLevel string) error {
	l := ctxzap.Extract(ctx)
	l.Debug("updating group page", zap.Int64("id", id), zap.String("access_level", accessLevel))

	args := []interface{}{accessLevel, id}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`UPDATE group_pages SET "accessLevel" = $1 WHERE "id"=$2`)

	_, err := c.db.Exec(ctx, sb.String(), args...)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteGroupPage(ctx context.Context, id int64) error {
	l := ctxzap.Extract(ctx)
	l.Debug("deleting group page", zap.Int64("id", id))

	args := []interface{}{id}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`DELETE FROM group_pages WHERE "id"=$1`)

	if _, err := c.db.Exec(ctx, sb.String(), args...); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetGroupFolderDefault(ctx context.Context, groupID int64, folderID int64) (*GroupFolderDefault, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group folder default", zap.Int64("group_id", groupID), zap.Int64("folder_id", folderID))

	args := []interface{}{groupID, folderID}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(``)

	var ret GroupFolderDefault
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) GetGroupResource(ctx context.Context, groupID int64, resourceID int64) (*GroupResource, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group resource", zap.Int64("group_id", groupID), zap.Int64("resource_id", resourceID))

	args := []interface{}{groupID, resourceID}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`SELECT "id", "accessLevel" from group_resources WHERE "groupId"=$1 AND "resourceId"=$2`)

	var ret GroupResource
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) GetGroupResourceFolderDefault(ctx context.Context, groupID int64, folderID int64) (*GroupResourceFolderDefault, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group resource folder default", zap.Int64("group_id", groupID), zap.Int64("folder_id", folderID))

	args := []interface{}{groupID, folderID}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`SELECT "id", "accessLevel" from group_resource_folder_defaults WHERE "groupId"=$1 AND "resourceFolderId"=$2`)

	var ret GroupResourceFolderDefault
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) ListGroupMembers(ctx context.Context, groupID int64, pager *Pager, skipDisabledUsers bool) ([]*GroupMember, string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("listing group members for group", zap.Int64("group_id", groupID))

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	args = append(args, groupID)

	sb := &strings.Builder{}
	if skipDisabledUsers {
		_, _ = sb.WriteString(`select user_groups.id as id, "userId", "groupId", "isAdmin", users.enabled from user_groups
		 INNER JOIN users ON users.id = "user_groups"."userId" WHERE "groupId"=$1 and users.enabled = true ORDER BY "id" `)
	} else {
		_, _ = sb.WriteString(`select "id", "userId", "groupId", "isAdmin" from user_groups WHERE "groupId"=$1 ORDER BY "id" `)
	}

	args = append(args, limit+1)
	_, _ = sb.WriteString(fmt.Sprintf("LIMIT $%d ", len(args)))

	if offset > 0 {
		args = append(args, offset)
		_, _ = sb.WriteString(fmt.Sprintf("OFFSET $%d", len(args)))
	}

	var ret []*GroupMember
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

func (c *Client) GetGroup(ctx context.Context, groupID int64) (*GroupModel, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group", zap.Int64("group_id", groupID))

	args := []interface{}{groupID}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "name", "organizationId", "universalAccess", "universalResourceAccess",
       						  "universalQueryLibraryAccess", "userListAccess", "auditLogAccess", "unpublishedReleaseAccess"
							  from groups WHERE "id"=$1`)

	var ret GroupModel
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) GetGroupByName(ctx context.Context, organizationID *int64, name string) (*GroupModel, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group by name", zap.Any("organization_id", organizationID), zap.String("name", name))

	args := []interface{}{organizationID, name}
	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "name", "organizationId", "universalAccess", "universalResourceAccess",
       						  "universalQueryLibraryAccess", "userListAccess", "auditLogAccess", "unpublishedReleaseAccess"
							  from groups WHERE "organizationId"=$1 AND "name"=$2`)

	var ret GroupModel
	err := pgxscan.Get(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Client) CreateGroup(ctx context.Context, organizationID *int64, name string) error {
	l := ctxzap.Extract(ctx)
	l.Debug("create group", zap.Any("organization_id", organizationID), zap.String("name", name))

	args := []interface{}{name, organizationID}

	if _, err := c.db.Exec(ctx, `INSERT INTO groups ("name", "organizationId", "createdAt", "updatedAt", "archivedAt", "usageAnalyticsAccess", "themeAccess", "unpublishedReleaseAccess", "accountDetailsAccess") VALUES ($1, $2,NOW(), NOW(), NULL, false, false, false, false)`, args...); err != nil {
		return err
	}

	return nil
}

func (c *Client) ListGroupsForOrg(ctx context.Context, orgID int64, pager *Pager) ([]*GroupModel, string, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("listing groups for org", zap.Int64("org_id", orgID))

	offset, limit, err := pager.Parse()
	if err != nil {
		return nil, "", err
	}
	var args []interface{}

	sb := &strings.Builder{}
	_, _ = sb.WriteString(`select "id", "name", "organizationId", "universalAccess", "universalResourceAccess",
       						  "universalQueryLibraryAccess", "userListAccess", "auditLogAccess", "unpublishedReleaseAccess"
							  from groups `)

	if orgID != 0 {
		args = append(args, orgID)
		_, _ = sb.WriteString(fmt.Sprintf(`WHERE "organizationId"=$%d `, len(args)))
	} else {
		_, _ = sb.WriteString(`WHERE "organizationId" IS NULL `)
	}

	_, _ = sb.WriteString(`ORDER BY "id" `)

	args = append(args, limit+1)
	_, _ = sb.WriteString(fmt.Sprintf("LIMIT $%d ", len(args)))

	if offset > 0 {
		args = append(args, offset)
		_, _ = sb.WriteString(fmt.Sprintf("OFFSET $%d", len(args)))
	}

	var ret []*GroupModel
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

func (c *Client) GetGroupMember(ctx context.Context, groupID, userID int64) (*GroupMember, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("getting group member", zap.Int64("group_id", groupID), zap.Int64("user_id", userID))

	var args []interface{}
	args = append(args, groupID)
	args = append(args, userID)

	sb := &strings.Builder{}
	_, err := sb.WriteString(`SELECT "id", "userId", "groupId", "isAdmin" FROM user_groups WHERE "groupId"=$1 AND "userId"=$2 LIMIT 1`)
	if err != nil {
		return nil, err
	}

	var ret []*GroupMember
	err = pgxscan.Select(ctx, c.db, &ret, sb.String(), args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if len(ret) == 0 {
		return nil, nil
	}

	return ret[0], nil
}

func (c *Client) AddGroupMember(ctx context.Context, groupID, userID int64, isAdmin bool) error {
	l := ctxzap.Extract(ctx)
	l.Debug("add user to group", zap.Int64("user_id", userID))

	args := []interface{}{groupID, userID, isAdmin}

	if _, err := c.db.Exec(ctx, `INSERT INTO user_groups ("groupId", "userId", "isAdmin", "createdAt", "updatedAt") VALUES ($1, $2, $3, NOW(), NOW())`, args...); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateGroupMember(ctx context.Context, groupID, userID int64, isAdmin bool) (*GroupMember, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("update user in group", zap.Int64("user_id", userID))

	args := []interface{}{groupID, userID, isAdmin}
	sb := &strings.Builder{}
	_, err := sb.WriteString(`UPDATE user_groups SET "isAdmin"=$3, "updatedAt" = NOW() WHERE "groupId"=$1 AND "userId"=$2`)
	if err != nil {
		return nil, err
	}

	if _, err := c.db.Exec(ctx, sb.String(), args...); err != nil {
		return nil, err
	}

	return c.GetGroupMember(ctx, groupID, userID)
}

func (c *Client) RemoveGroupMember(ctx context.Context, groupID, userID int64) error {
	l := ctxzap.Extract(ctx)
	l.Debug("remove user from group", zap.Int64("user_id", userID))

	args := []interface{}{groupID, userID}
	sb := &strings.Builder{}
	_, err := sb.WriteString(`DELETE FROM user_groups WHERE "groupId"=$1 AND "userId"=$2`)
	if err != nil {
		return err
	}

	if _, err := c.db.Exec(ctx, sb.String(), args...); err != nil {
		return err
	}

	return nil
}
