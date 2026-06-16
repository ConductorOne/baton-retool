#!/usr/bin/env bash
# Exercise the account lifecycle end-to-end against the local Retool stack:
#   create -> verify -> disable_user action -> verify -> enable_user action -> verify
#   -> duplicate disable (idempotency).
#
# Retool's REST API has no hard delete, so the connector models deprovisioning as the
# reversible enable_user/disable_user actions (PATCH /active) instead of Delete.
#
# Requires (set by the workflow): BATON_CONNECTION_STRING, BATON_RETOOL_API_BASE_URL,
# BATON_RETOOL_API_TOKEN. Run from the repo root with ./baton-retool already built.
set -euo pipefail

CONNECTOR="${CONNECTOR:-./baton-retool}"
BASE_URL="${BATON_RETOOL_API_BASE_URL:?BATON_RETOOL_API_BASE_URL must be set}"
TOKEN="${BATON_RETOOL_API_TOKEN:?BATON_RETOOL_API_TOKEN must be set}"
EMAIL="baton.ci.provision@example.com"
C1Z="$(mktemp -t prov-XXXX.c1z)"

fail() { echo "FAIL: $*" >&2; exit 1; }
lookup() { curl -sf "$BASE_URL/api/v2/users?email=$EMAIL" -H "Authorization: Bearer $TOKEN"; }

echo "==> create account $EMAIL"
"$CONNECTOR" --provisioning --file "$C1Z" \
  --create-account-login="$EMAIL" \
  --create-account-email="$EMAIL" \
  --create-account-profile="{\"email\":\"$EMAIL\",\"first_name\":\"Baton\",\"last_name\":\"CI\",\"user_type\":\"default\"}"

echo "==> verify the account exists and is active"
RESP="$(lookup)"
echo "$RESP" | grep -q '"email":"'"$EMAIL"'"' || fail "created account not found via REST"
echo "$RESP" | grep -q '"active":true' || fail "created account is not active"
LID="$(echo "$RESP" | sed -E 's/.*"legacy_id":([0-9]+).*/\1/')"
[ -n "$LID" ] || fail "could not resolve legacy_id"
echo "    legacy_id=$LID"

# user_id is a ResourceIdField, so it must be passed as a {resource_type_id, resource_id}
# struct (mirrors the C1 resource picker), not a bare string.
USER_ARG="{\"user_id\":{\"resource_type_id\":\"user\",\"resource_id\":\"u$LID\"}}"

echo "==> disable_user action on u$LID"
"$CONNECTOR" --provisioning --file "$C1Z" \
  --invoke-action disable_user --invoke-action-args "$USER_ARG"

echo "==> verify the account is deactivated"
echo "$(lookup)" | grep -q '"active":false' || fail "account was not deactivated by disable_user"

echo "==> enable_user action on u$LID"
"$CONNECTOR" --provisioning --file "$C1Z" \
  --invoke-action enable_user --invoke-action-args "$USER_ARG"

echo "==> verify the account is active again"
echo "$(lookup)" | grep -q '"active":true' || fail "account was not reactivated by enable_user"

echo "==> duplicate disable must be idempotent (success, not error)"
"$CONNECTOR" --provisioning --file "$C1Z" \
  --invoke-action disable_user --invoke-action-args "$USER_ARG"
"$CONNECTOR" --provisioning --file "$C1Z" \
  --invoke-action disable_user --invoke-action-args "$USER_ARG"
echo "$(lookup)" | grep -q '"active":false' || fail "account is not deactivated after duplicate disable"

echo "PASS: account create -> disable -> enable -> dup-disable all succeeded"
