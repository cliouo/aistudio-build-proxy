#!/usr/bin/env bash
set -e
ARGS=("--url" "$TARGET_URL" "--cookie" "$COOKIE_FILE")
[ -n "$HTTP_PROXY" ] && ARGS+=("--proxy" "$HTTP_PROXY")
[ "$HEADLESS" != "false" ] && ARGS+=("--headless")
exec python /app/runner.py "${ARGS[@]}"
