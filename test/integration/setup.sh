#!/usr/bin/env bash
# Wait for the local Retool stack to be ready, create the initial admin/org, and
# mint a REST API token (with users:read + users:write) for the integration test.
#
# Diagnostics go to stderr; the ONLY thing printed to stdout is the API token, so
# the workflow can capture it with: TOKEN=$(setup.sh)
#
# Retool stores personal access tokens as hashedKey = sha256(token), so we can
# mint one deterministically with a SQL insert — no UI required.
set -euo pipefail

BASE_URL="${RETOOL_BASE_URL:-http://localhost:3000}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-BatonCITest123!}"
API_TOKEN="${API_TOKEN:-retool_ci$(openssl rand -hex 12)}"
COMPOSE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { echo "[setup] $*" >&2; }
psql_q() { docker compose -f "$COMPOSE_DIR/compose.yaml" exec -T postgres psql -U retool_internal_user -d hammerhead_production -tAc "$1"; }

# 1. Wait for the api (and therefore the jobs-runner migrations) to be healthy.
log "waiting for Retool api to become healthy at $BASE_URL ..."
for i in $(seq 1 120); do
  if curl -sf "$BASE_URL/api/checkHealth" 2>/dev/null | grep -q '"status":"HEALTHY"'; then
    log "api healthy after ~$((i * 5))s"
    break
  fi
  if [ "$i" -eq 120 ]; then
    log "ERROR: api did not become healthy in time"
    exit 1
  fi
  sleep 5
done

# 2. Create the initial admin + org (first-run signup needs no auth). Idempotent:
#    if an admin already exists (re-run against a seeded DB), skip signup.
if [ "$(psql_q "SELECT count(*) FROM users WHERE \"organizationId\"=1;")" = "0" ]; then
  log "creating admin account $ADMIN_EMAIL ..."
  curl -sf -X POST "$BASE_URL/api/signup" -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\",\"name\":\"Baton CI\",\"planType\":\"free\"}" \
    >/dev/null || { log "ERROR: signup failed"; exit 1; }
else
  log "admin/org already present, skipping signup"
fi

# 3. Mint a REST API token via SQL (hashedKey = sha256(token); scope users:read+write).
log "minting REST API token ..."
HASH=$(printf '%s' "$API_TOKEN" | sha256sum | cut -d' ' -f1)
LAST4="${API_TOKEN: -4}"
psql_q "INSERT INTO personal_access_tokens (id,label,\"hashedKey\",\"organizationId\",\"userId\",revoked,scope,\"createdAt\",\"updatedAt\",last4) VALUES (gen_random_uuid(),'baton-ci','$HASH',1,1,false,'[\"users:read\",\"users:write\"]'::jsonb,now(),now(),'$LAST4');" >/dev/null

# 4. Sanity-check the token authenticates.
if ! curl -sf "$BASE_URL/api/v2/users?limit=1" -H "Authorization: Bearer $API_TOKEN" | grep -q '"success":true'; then
  log "ERROR: minted token failed to authenticate"
  exit 1
fi
log "token authenticates OK"

# stdout: the token only.
echo "$API_TOKEN"
