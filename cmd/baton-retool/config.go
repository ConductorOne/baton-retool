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
)

var configurationFields = []field.SchemaField{
	ConnectionString,
	SkipPages,
	SkipResources,
	SkipDisabledUsers,
}

var configRelations = []field.SchemaFieldRelationship{}

var configuration = field.NewConfiguration(configurationFields, field.WithConstraints(configRelations...))
