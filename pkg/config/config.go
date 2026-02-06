package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ConnectionString = field.StringField(
		"connection-string",
		field.WithRequired(true),
		field.WithDescription("The connection string for connecting to retool database"),
		field.WithIsSecret(true),
		field.WithDisplayName("Connection String"),
	)
	SkipPages = field.BoolField(
		"skip-pages",
		field.WithDescription("Skip syncing pages"),
		field.WithDisplayName("Skip Pages"),
	)
	SkipResources = field.BoolField(
		"skip-resources",
		field.WithDescription("Skip syncing resources"),
		field.WithDisplayName("Skip Resources"),
	)
	SkipDisabledUsers = field.BoolField(
		"skip-disabled-users",
		field.WithDescription("Skip syncing disabled users"),
		field.WithDisplayName("Skip Disabled Users"),
	)

	// FieldRelationships defines relationships between the fields.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Configuration = field.NewConfiguration([]field.SchemaField{
	ConnectionString,
	SkipPages,
	SkipResources,
	SkipDisabledUsers,
}, field.WithConstraints(FieldRelationships...))
