#!/usr/bin/env bash
set -e
export PYTHONUNBUFFERED=1
PORT=${CFX_PORT:-38835}
LOG=/tmp/camoufox.log
if [ "$PORT" = "0" ]; then
  python -m camoufox server --virtual-display | tee "$LOG" &
  while ! grep -q "Websocket endpoint:" "$LOG"; do sleep 0.2; done
  grep "Websocket endpoint:" "$LOG" | awk '{print $3}' > /tmp/endpoint.txt
else
  python -m camoufox server --virtual-display --port "$PORT" &
  echo "ws://localhost:${PORT}" > /tmp/endpoint.txt
fi
wait
