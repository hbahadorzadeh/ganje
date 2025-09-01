#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

export COMPOSE_PROJECT_NAME=ganje_it

JWT_SECRET="dev-jwt-secret"
ADMIN_REALMS="admins"
DEV_REALMS="developers"

BACKEND_URL="http://localhost:8080"
HEALTH_URL="$BACKEND_URL/health"
API_BASE="$BACKEND_URL/api/v1"

ADMIN_TOKEN=""
DEV_TOKEN=""

function cleanup() {
  echo "\n[cleanup] docker compose down"
  docker compose down -v || true
}
trap cleanup EXIT

echo "[compose] bringing up services"
docker compose up -d --build

# wait for backend
"$ROOT_DIR/scripts/it/wait-for-http.sh" "$HEALTH_URL" 180 3

echo "[auth] generating tokens"
ADMIN_TOKEN=$(go run ./scripts/it/jwt.go "$JWT_SECRET" "ci-admin" "ci-admin@example.com" "$ADMIN_REALMS")
DEV_TOKEN=$(go run ./scripts/it/jwt.go "$JWT_SECRET" "ci-dev" "ci-dev@example.com" "$DEV_REALMS")

echo "[api] create repositories (generic-local, maven-local)"
create_repo() {
  local name=$1; shift
  local type=$1; shift
  local art=$1; shift
  curl -sfS -X POST "$API_BASE/repositories" \
    -H "Authorization: $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"type\":\"$type\",\"artifact_type\":\"$art\"}"
}

# best-effort; ignore if exists
create_repo generic-local local generic || true
create_repo maven-local local maven || true

# generic push/pull/delete
GEN_PATH="dir1/dir2/hello.txt"
GEN_DATA="hello-world-$(date +%s)"

echo "[generic] PUT $GEN_PATH"
# send data as binary body
curl -sfS -X PUT "$BACKEND_URL/generic-local/$GEN_PATH" \
  -H "Authorization: $DEV_TOKEN" \
  --data-binary "$GEN_DATA" \
  -o /dev/null

echo "[generic] GET $GEN_PATH and verify"
RESP=$(curl -sfS "$BACKEND_URL/generic-local/$GEN_PATH" -H "Authorization: $DEV_TOKEN")
if [[ "$RESP" != "$GEN_DATA" ]]; then
  echo "generic content mismatch"
  exit 1
fi

echo "[generic] DELETE $GEN_PATH"
curl -sfS -X DELETE "$BACKEND_URL/generic-local/$GEN_PATH" -H "Authorization: $DEV_TOKEN" -o /dev/null

# maven push/pull
MAVEN_PATH="com/example/app/1.0.0/app-1.0.0.jar"
MAVEN_DATA="dummy-jar-$(date +%s)"

echo "[maven] PUT $MAVEN_PATH"
curl -sfS -X PUT "$BACKEND_URL/maven-local/$MAVEN_PATH" \
  -H "Authorization: $DEV_TOKEN" \
  --data-binary "$MAVEN_DATA" \
  -o /dev/null

echo "[maven] GET $MAVEN_PATH and verify"
RESP2=$(curl -sfS "$BACKEND_URL/maven-local/$MAVEN_PATH" -H "Authorization: $DEV_TOKEN")
if [[ "$RESP2" != "$MAVEN_DATA" ]]; then
  echo "maven content mismatch"
  exit 1
fi

echo "[done] integration tests passed"
