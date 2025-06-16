#!/usr/bin/env bash
set -e
export PYTHONUNBUFFERED=1

if [[ -n "$USE_XVFB" ]]; then
  xvfb-run --server-arg="-screen 0 1920x1080x24" \
    python -m camoufox server --virtual-display --port ${CFX_PORT:-38835}
else
  python -m camoufox server --headless --port ${CFX_PORT:-38835}
fi
