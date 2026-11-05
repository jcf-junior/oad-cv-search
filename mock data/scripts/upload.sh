#!/usr/bin/env bash
set -u

API_URL="${1:-http://localhost:8080}"
DIR="${2:-$(dirname "$0")/../pdfs}"

ok=0
failed=0

for file in "$DIR"/*.pdf; do
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/candidates" -F "file=@$file")
  if [ "$status" = "201" ]; then
    ok=$((ok + 1))
    echo "ok      $(basename "$file")"
  else
    failed=$((failed + 1))
    echo "$status     $(basename "$file")"
  fi
done

echo "uploaded: $ok, failed: $failed"
