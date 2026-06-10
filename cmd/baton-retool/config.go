package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ConnectionString = field.StringField(
		"connection-string",
		field.WithRequired(true),
		field.WithDescription("The connection string for connecting to retool database"),
	)
	SkipPages = field.BoolField(
		"skip-pages",
		field.WithDescription("Skip syncing pages"),
	)
	SkipResources = field.BoolField(
		"skip-resources",
		field.WithDescription("Skip syncing resources"),
	)
	SkipDisabledUsers = field.BoolField(
		"skip-disabled-users",
		field.WithDescription("Skip syncing disabled users"),
	)
	// REST surface for account provisioning/deprovisioning (CXH-1585). Optional:
	// sync-only deployments keep working without these; the lifecycle handlers fail
	// fast with a clear error when they are absent.
	RetoolAPIBaseURL = field.StringField(
		"retool-api-base-url",
		field.WithDescription("Base URL of the Retool REST API, e.g. https://<org>.retool.com. Required only for account provisioning/deprovisioning."),
	)
	RetoolAPIToken = field.StringField(
		"retool-api-token",
		field.WithDescription("Retool API token with users:read + users:write. Required only for account provisioning/deprovisioning."),
		field.WithIsSecret(true),
	)
)

var configurationFields = []field.SchemaField{
	ConnectionString,
	SkipPages,
	SkipResources,
	SkipDisabledUsers,
	RetoolAPIBaseURL,
	RetoolAPIToken,
}

// retool-api-base-url and retool-api-token are both-or-neither.
var configRelations = []field.SchemaFieldRelationship{
	field.FieldsRequiredTogether(RetoolAPIBaseURL, RetoolAPIToken),
}

var configuration = field.NewConfiguration(configurationFields, field.WithConstraints(configRelations...))
