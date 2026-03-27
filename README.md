# baton-retool
`baton-retool` is a connector for [Retool](https://retool.com/) built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It connects to the Retool API v2 and syncs data about users, groups, organizations, apps (pages), and resources.

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## Setup
1. Generate a Retool API access token:
   - Go to **Settings > API** in your Retool instance
   - Create a new API token with the following scopes:
     - **Users > Read** (to list and get users)
     - **Users > Write** (for account provisioning - creating users)
     - **Groups > Read** (to list groups and members)
     - **Groups > Write** (for group membership management)
     - **Permissions > Read** (to read app and resource permissions)
     - **Permissions > Write** (for granting/revoking permissions)

2. Run the connector with your API token and Retool instance URL:
```bash
BATON_API_TOKEN="your-retool-api-token" BATON_API_URL="https://your-org.retool.com" baton-retool
```

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-retool
BATON_API_TOKEN="your-retool-api-token" BATON_API_URL="https://your-org.retool.com" baton-retool
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_API_TOKEN="your-retool-api-token" -e BATON_API_URL="https://your-org.retool.com" ghcr.io/conductorone/baton-retool:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-retool/cmd/baton-retool@main

BATON_API_TOKEN="your-retool-api-token" BATON_API_URL="https://your-org.retool.com" baton-retool
baton resources
```

# Data Model

`baton-retool` pulls down information about the following Retool resources:
- Users
- Groups
- Organizations
- Apps (Pages)
- Resources

## Account Provisioning

The connector supports creating new user accounts in Retool via the API. When provisioning is enabled (`-p` flag), the connector can create new users with their email, first name, and last name.

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
      --api-token string             required: The Retool API access token ($BATON_API_TOKEN)
      --api-url string               required: The base URL of the Retool instance ($BATON_API_URL)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-retool
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --skip-disabled-users          Skip syncing disabled/inactive users ($BATON_SKIP_DISABLED_USERS)
      --skip-full-sync               This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --skip-pages                   Skip syncing apps/pages ($BATON_SKIP_PAGES)
      --skip-resources               Skip syncing resources ($BATON_SKIP_RESOURCES)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-retool

Use "baton-retool [command] --help" for more information about a command.
```
