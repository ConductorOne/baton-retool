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
		DisplayName: "retool",
		Description: "Retool connector using API v2",
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
		newOrgSyncer(ctx, c.client, c.skipPages, c.skipResources, c.skipDisabledUsers),
		newUserSyncer(ctx, c.client, c.skipDisabledUsers),
		newGroupSyncer(ctx, c.client, c.skipDisabledUsers),
	}

	if !c.skipPages {
		syncers = append(syncers, newPageSyncer(c.client, c.skipDisabledUsers))
	}

	if !c.skipResources {
		syncers = append(syncers, newResourceSyncer(ctx, c.client))
	}

	return syncers
}

// CreateAccount implements the AccountManager interface to provision new user accounts.
func (c *ConnectorImpl) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.CredentialOptions) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	if accountInfo == nil {
		return nil, nil, nil, fmt.Errorf("baton-retool: account info is required to create user")
	}

	var email string
	emails := accountInfo.GetEmails()
	for _, e := range emails {
		if e.GetIsPrimary() {
			email = e.GetAddress()
			break
		}
		if email == "" {
			email = e.GetAddress()
		}
	}

	if email == "" {
		email = accountInfo.GetLogin()
	}

	if email == "" {
		return nil, nil, nil, fmt.Errorf("baton-retool: email is required to create user")
	}

	var firstName, lastName string
	if accountInfo.GetProfile() != nil {
		fields := accountInfo.GetProfile().GetFields()
		if v, ok := fields["first_name"]; ok {
			firstName = v.GetStringValue()
		}
		if v, ok := fields["last_name"]; ok {
			lastName = v.GetStringValue()
		}
	}

	user, err := c.client.CreateUser(ctx, &client.CreateUserRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Active:    true,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("baton-retool: failed to create user account: %w", err)
	}

	_ = user

	result := &v2.CreateAccountResponse_SuccessResult{
		IsCreateAccountResult: true,
	}

	return result, nil, nil, nil
}

// CreateAccountCapabilityDetails returns the details of the account provisioning capability.
func (c *ConnectorImpl) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{}, nil, nil
}

func New(ctx context.Context, baseURL string, apiToken string, skipPages bool, skipResources bool, skipDisabledUsers bool) (*ConnectorImpl, error) {
	c, err := client.New(ctx, baseURL, apiToken)
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
