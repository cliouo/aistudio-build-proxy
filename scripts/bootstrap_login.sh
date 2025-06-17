#!/usr/bin/env bash
set -e
COOKIE=${COOKIE_FILE:-./data/cookie.json}
URL=${TARGET_URL:?need url}
python python/runner.py --url "$URL" --cookie "$COOKIE" --bootstrap
