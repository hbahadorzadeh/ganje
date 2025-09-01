#!/usr/bin/env bash
set -euo pipefail

URL=${1:-http://localhost:8080/health}
TIMEOUT=${2:-120}
SLEEP=${3:-3}

end=$((SECONDS+TIMEOUT))
echo "Waiting for $URL ... (timeout ${TIMEOUT}s)"
while [ $SECONDS -lt $end ]; do
  code=$(curl -sk -o /dev/null -w "%{http_code}" "$URL" || true)
  if [ "$code" = "200" ]; then
    echo "Service is up"
    exit 0
  fi
  sleep "$SLEEP"
  echo -n "."
done

echo "Timed out waiting for $URL"
exit 1
