# Baton Retool - Connector Documentation

This document provides information needed to set up and use the connector.

## Connector Capabilities

### 1. What resources does the connector sync?

| Resource | Description |
|----------|-------------|
| **Organization** | The Retool org/tenant; parent scope for all other resources. |
| **User** | Retool user accounts (principals) with email, name, and enabled/disabled status. |
| **Group** | Permission groups; users hold membership (and optional group-admin) grants. |
| **Page (App)** | Retool apps; groups are granted an access level (read/write/own). |
| **Resource** | Retool data resources (databases, APIs); synced for visibility. |

### 2. Can the connector provision any resources? If so, which ones?

Yes.

| Resource | Grant | Revoke | Create | Delete |
|----------|-------|--------|--------|--------|
| **Group membership** | ✅ Adds the user to the group (Postgres `user_groups`) | ✅ Removes the user from the group | - | - |
| **Page (App) access** | ✅ Grants the group an access level on the page | ✅ Removes the group's access | - | - |
| **Account (User)** | - | - | ✅ Creates/invites a Retool user via the REST API | ✅ Deprovisions the user via the REST API |

**Important behavioral notes:**
- **Sync and group/page provisioning** run entirely against the Retool **Postgres database** (the `connection-string`). They do not require the REST API.
- **Account Create/Delete** use the Retool **REST API** (`/api/v2/users`) and require `retool-api-base-url` + `retool-api-token`. These are optional config; when absent, the account-lifecycle handlers fail fast with a clear "REST API not configured" error while sync and grant/revoke keep working.
- **Account "Delete" is a soft deactivation, not a hard delete.** Retool's REST API has no hard-delete endpoint — `DELETE /api/v2/users/{id}` sets the user to disabled (blocks sign-in), retains group memberships, and is reversible. Re-deleting an already-deactivated user and deleting an unknown user are both treated as success (idempotent).
- The connector resolves the synced `user:<int64>` (Postgres `id`, exposed as `legacy_id` over REST) to the REST `sid` (`user_<uuid>`) via a direct Postgres lookup — no email-based matching.
- There is **no enable/disable connector action** (the connector's SDK version predates the action framework); deprovisioning is exposed via account Delete.

## Connector Credentials

### 1. What credentials or information are needed to set up the connector?

| Credential | Required | Description |
|------------|----------|-------------|
| **Connection string** | Yes | Postgres DSN for the Retool database (`user=… password=… host=… port=5432 dbname=hammerhead_production`). Used for sync and group/page provisioning. |
| **Retool API base URL** | No* | Retool base URL, e.g. `https://<org>.retool.com`. Required only for account provisioning. |
| **Retool API token** | No* | Retool API token with `users:read` + `users:write` scopes. Required only for account provisioning. |

\* `retool-api-base-url` and `retool-api-token` are required together (both or neither).

### 2. How are these credentials obtained?

- **Connection string:** Connect to the Retool Postgres database and create a dedicated user (`CREATE USER baton …`) with the SELECT/INSERT/UPDATE/DELETE grants listed in the repo `README.md`, then compose the DSN.
- **API token:** In Retool, go to **Settings → API**, create a token, and grant it the `users:read` and `users:write` scopes. Note the org base URL.

## Additional Notes

### Retool Plan Requirements

- **Direct database access** to Retool's primary Postgres DB is generally a **self-hosted Retool** capability (or a managed/peered database you can reach from where the connector runs). This is required for all sync and group/page provisioning.
- **REST API access** (API tokens) may be gated behind a specific Retool plan/tier. Account provisioning/deprovisioning is only available where the REST API and a `users:read`+`users:write` token are available.

### API Documentation Links

- [Retool REST API reference](https://docs.retool.com/reference/api) — user-management endpoints (`/api/v2/users`).
- [Retool API authentication](https://docs.retool.com/reference/api/authentication) — token scopes.
