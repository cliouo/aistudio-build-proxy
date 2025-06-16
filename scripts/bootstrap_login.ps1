param(
  [string]$BrowserUrl = $Env:BROWSER_URL,
  [int]$Port = $Env:CFX_PORT -as [int] -or 38835
)

Start-Process -PassThru python "-m camoufox server --port $Port" | Out-Null
$env:CAMOUFOX_WS = "ws://localhost:$Port"

Write-Host "👉 请在弹出的 Firefox 中完成 Google 登录和 2FA，然后按回车继续 …"
& go run ./cmd/agent --bootstrap --ws-endpoint $env:CAMOUFOX_WS --url $BrowserUrl
