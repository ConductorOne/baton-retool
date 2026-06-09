#!/usr/bin/env bash
# Exercise account provisioning end-to-end against the local Retool stack:
#   create -> verify -> delete (deactivate) -> verify -> duplicate-delete (idempotency).
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

echo "==> delete (deactivate) u$LID"
"$CONNECTOR" --provisioning --file "$C1Z" --delete-resource="u$LID" --delete-resource-type=user

echo "==> verify the account is deactivated (Retool delete is a soft-disable)"
echo "$(lookup)" | grep -q '"active":false' || fail "account was not deactivated after delete"

echo "==> duplicate delete must be idempotent (success, not error)"
"$CONNECTOR" --provisioning --file "$C1Z" --delete-resource="u$LID" --delete-resource-type=user

echo "PASS: account provisioning create -> delete -> dup-delete all succeeded"
