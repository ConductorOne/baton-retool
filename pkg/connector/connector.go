package connector

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/conductorone/baton-retool/pkg/client"
)

func titleCase(s string) string {
	titleCaser := cases.Title(language.English)

	return titleCaser.String(s)
}

type ConnectorImpl struct {
	client            *client.Client
	skipPages         bool
	skipResources     bool
	skipDisabledUsers bool
}

func (c *ConnectorImpl) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Retool",
		Description: "Retool connector",
		// Drives the form ConductorOne renders when provisioning a new account. Field
		// keys must match what userSyncer.CreateAccount reads from AccountInfo.Profile.
		AccountCreationSchema: &v2.ConnectorAccountCreationSchema{
			FieldMap: map[string]*v2.ConnectorAccountCreationSchema_Field{
				"email": {
					DisplayName: "Email",
					Required:    true,
					Description: "Email address of the user.",
					Placeholder: "email@example.com",
					Order:       1,
					Field:       &v2.ConnectorAccountCreationSchema_Field_StringField{StringField: &v2.ConnectorAccountCreationSchema_StringField{}},
				},
				"first_name": {
					DisplayName: "First Name",
					Required:    true,
					Description: "First name of the user.",
					Placeholder: "First Name",
					Order:       2,
					Field:       &v2.ConnectorAccountCreationSchema_Field_StringField{StringField: &v2.ConnectorAccountCreationSchema_StringField{}},
				},
				"last_name": {
					DisplayName: "Last Name",
					Required:    true,
					Description: "Last name of the user.",
					Placeholder: "Last Name",
					Order:       3,
					Field:       &v2.ConnectorAccountCreationSchema_Field_StringField{StringField: &v2.ConnectorAccountCreationSchema_StringField{}},
				},
				"user_type": {
					DisplayName: "User Type",
					Required:    false,
					Description: "Retool user type: \"default\" (full platform user, billable), \"mobile\", or \"embed\". Defaults to \"default\".",
					Placeholder: "default",
					Order:       4,
					Field:       &v2.ConnectorAccountCreationSchema_Field_StringField{StringField: &v2.ConnectorAccountCreationSchema_StringField{}},
				},
			},
		},
	}, nil
}

func (c *ConnectorImpl) Validate(ctx context.Context) (annotations.Annotations, error) {
	err := c.client.ValidateConnection(ctx)
	if err != nil {
		return nil, err
	}

	// Probe the REST surface only when it is configured (sync-only deployments skip it).
	if c.client.RESTEnabled() {
		annos, err := c.client.ValidateREST(ctx)
		if err != nil {
			return nil, fmt.Errorf("baton-retool: REST API validation failed: %w", err)
		}
		return annos, nil
	}

	return nil, nil
}

func (c *ConnectorImpl) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

func (c *ConnectorImpl) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	syncers := []connectorbuilder.ResourceSyncer{
		newOrgSyncer(ctx, c.client, c.skipPages, c.skipResources, c.skipDisabledUsers),
		newUserSyncer(ctx, c.client, c.skipDisabledUsers),
		newGroupSyncer(ctx, c.client, c.skipDisabledUsers),
	}

	if !c.skipPages {
		syncers = append(syncers, newPageSyncer(c.client, c.skipDisabledUsers))
	}

	if !c.skipResources {
		syncers = append(syncers, newResourceSyncer(ctx, c.client, c.skipDisabledUsers))
	}

	return syncers
}

func New(ctx context.Context, dsn string, skipPages bool, skipResources bool, skipDisabledUsers bool, apiBaseURL string, apiToken string) (*ConnectorImpl, error) {
	c, err := client.New(ctx, dsn, apiBaseURL, apiToken)
	if err != nil {
		return nil, err
	}

	return &ConnectorImpl{
		client:            c,
		skipPages:         skipPages,
		skipResources:     skipResources,
		skipDisabledUsers: skipDisabledUsers,
	}, nil
}
