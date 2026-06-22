package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	APIToken = field.StringField(
		"api-token",
		field.WithRequired(true),
		field.WithDescription("The Retool API access token (Bearer token)"),
	)
	APIURL = field.StringField(
		"api-url",
		field.WithRequired(true),
		field.WithDescription("The base URL of the Retool instance (e.g., https://myorg.retool.com)"),
	)
	SkipPages = field.BoolField(
		"skip-pages",
		field.WithDescription("Skip syncing apps/pages"),
	)
	SkipResources = field.BoolField(
		"skip-resources",
		field.WithDescription("Skip syncing resources"),
	)
	SkipDisabledUsers = field.BoolField(
		"skip-disabled-users",
		field.WithDescription("Skip syncing disabled/inactive users"),
	)
)

var configurationFields = []field.SchemaField{
	APIToken,
	APIURL,
	SkipPages,
	SkipResources,
	SkipDisabledUsers,
}

var configRelations = []field.SchemaFieldRelationship{}

var configuration = field.NewConfiguration(configurationFields, configRelations...)
