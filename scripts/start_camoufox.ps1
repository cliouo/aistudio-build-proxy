param(
  [int]$Port = $Env:CFX_PORT -as [int] -or 38835,
  [switch]$UseXvfb
)

$log = Join-Path $PSScriptRoot 'camoufox.log'
if (Test-Path $log) { Remove-Item $log }

if ($Port -ne 0) {
  python -m camoufox server --virtual-display --port $Port 2>&1 |
    Tee-Object -FilePath $log
  "ws://localhost:$Port" | Out-File "$Env:TEMP\endpoint.txt" -Encoding ascii
} else {
  python -m camoufox server --virtual-display 2>&1 |
    Tee-Object -FilePath $log
  (Get-Content $log -Wait | Select-String -Pattern 'ws://localhost:\d+/\S+' -First 1).Line |
    Out-File "$Env:TEMP\endpoint.txt" -Encoding ascii
}
