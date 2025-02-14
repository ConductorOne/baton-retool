# baton-retool
`baton-retool` is a connector for [Retool](https://retool.com/) built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It connects directly to your primary Retool Postgres database and syncs data about users, groups, organizations, pages, and resources.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## Setup
1. While connected to the Retool database, create a new user for the connector to connect to Postgres as. Be sure to create and save the secure password for this user:
```postgresql
CREATE USER baton WITH PASSWORD 'secure-password';
```
2. Grant your new role the privileges required by the connector for inspecting retool privileges.
```postgresql
GRANT SELECT ("id", "name", "organizationId", "universalAccess", "universalResourceAccess", "universalQueryLibraryAccess",
              "userListAccess", "auditLogAccess", "unpublishedReleaseAccess") ON groups TO baton;
GRANT SELECT, INSERT, UPDATE ("id", "accessLevel"), DELETE ON group_pages TO baton;
GRANT SELECT, INSERT, UPDATE ("id", "accessLevel") ON group_folder_defaults TO baton;
GRANT SELECT, INSERT, UPDATE ("id", "accessLevel") on group_resources TO baton;
GRANT SELECT, INSERT, UPDATE ("id", "accessLevel") on group_resource_folder_defaults TO baton;
GRANT SELECT ("id", "name") ON organizations TO baton;
GRANT SELECT ("id", "name", "organizationId", "folderId", "photoUrl", "description", "deletedAt") ON pages TO baton;
GRANT SELECT ("id", "name", "organizationId", "type", "displayName", "environmentId", "resourceFolderId") ON resources TO baton;
GRANT SELECT ("id", "email", "firstName", "lastName", "profilePhotoUrl", "userName", "enabled", "lastLoggedIn", "organizationId") ON users TO baton;
GRANT SELECT, INSERT, UPDATE, DELETE ("id", "userId", "groupId", "isAdmin", "updatedAt") ON user_groups TO baton;
GRANT USAGE, SELECT ON SEQUENCE user_groups_id_seq TO baton;
GRANT DELETE ON user_groups TO baton;
```

3. Run the connector with the proper connection string. For example if you created a new `baton` user with the password `baton`, it may look like this:
```bash
BATON_CONNECTION_STRING="user=baton password=baton host=localhost port=5432 dbname=hammerhead_production" baton-retool
```

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-retool
baton-retool
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_CONNECTION_STRING="user=baton password=baton host=localhost port=5432 dbname=hammerhead_production" ghcr.io/conductorone/baton-retool:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-retool/cmd/baton-retool@main

BATON_CONNECTION_STRING="user=baton password=baton host=localhost port=5432 dbname=hammerhead_production" baton-retool
baton resources
```

# Data Model

`baton-retool` pulls down information about the following Retool resources:
- Users
- Groups
- Organizations
- Pages
- Resources

# Contributing, Support, and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-retool` Command Line Usage

```
baton-retool

Usage:
  baton-retool [flags]
  baton-retool [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string           The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string       The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
      --connection-string string   required: The connection string for connecting to retool database ($BATON_CONNECTION_STRING)
  -f, --file string                The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                       help for baton-retool
      --log-format string          The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string           The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning               This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --skip-disabled-users        Skip syncing disabled users ($BATON_SKIP_DISABLED_USERS)
      --skip-full-sync             This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --skip-pages                 Skip syncing pages ($BATON_SKIP_PAGES)
      --skip-resources             Skip syncing resources ($BATON_SKIP_RESOURCES)
      --ticketing                  This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                    version for baton-retool

Use "baton-retool [command] --help" for more information about a command.
```