#!/usr/bin/env bash
set -e

python -m camoufox server &
PID=$!
trap 'kill $PID' EXIT

go run ./cmd/agent --bootstrap --ws-endpoint ws://localhost:38835
