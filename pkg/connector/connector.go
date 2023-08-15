package connector

import (
	"context"
	"io"

	"github.com/conductorone/baton-retool/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func titleCase(s string) string {
	titleCaser := cases.Title(language.English)

	return titleCaser.String(s)
}

type ConnectorImpl struct {
	client *client.Client
}

func (c *ConnectorImpl) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "retool",
	}, nil
}

func (c *ConnectorImpl) Validate(ctx context.Context) (annotations.Annotations, error) {
	err := c.client.ValidateConnection(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *ConnectorImpl) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

func (c *ConnectorImpl) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	syncers := []connectorbuilder.ResourceSyncer{
		newOrgSyncer(ctx, c.client),
		newUserSyncer(ctx, c.client),
		newGroupSyncer(ctx, c.client),
		newPageSyncer(ctx, c.client),
		newResourceSyncer(ctx, c.client),
	}

	return syncers
}

func New(ctx context.Context, dsn string) (*ConnectorImpl, error) {
	c, err := client.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &ConnectorImpl{
		client: c,
	}, nil
}
