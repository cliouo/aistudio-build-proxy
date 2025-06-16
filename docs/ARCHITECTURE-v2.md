# Camoufox Agent Architecture v2.0

This version documents the latest changes to the agent and server setup.

## 1. Port handling and WebSocket discovery
- `scripts/start_camoufox.sh` and `start_camoufox.ps1` write the detected WebSocket endpoint to `/tmp/endpoint.txt`.
- Set `CFX_PORT=0` to use a random port; otherwise a fixed port (default `38835`) is used.
- Agents read `CAMOUFOX_WS` from the environment or `/tmp/endpoint.txt` automatically.

## 2. Agent variables
- `BROWSER_URL` is the full URL kept open in the headless browser.
- `AUTH_TOKEN` is only used by the WebSocket gateway and will no longer be appended to `BROWSER_URL`.

## 3. Bootstrap login
- `scripts/bootstrap_login.(sh|ps1)` starts Camoufox in GUI mode and runs the agent with `--bootstrap`.
- Complete Google login in the displayed Firefox window and press <Enter> to persist cookies to `./data/google.json`.

## 4. Windows support
- PowerShell equivalents `start_camoufox.ps1` and `bootstrap_login.ps1` provide the same behaviour on Windows.
