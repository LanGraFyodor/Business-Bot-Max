$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
if (Test-Path -LiteralPath '.env') {
    foreach ($line in Get-Content -LiteralPath '.env') {
        if ($line -match '^\s*([A-Z_][A-Z0-9_]*)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2].Trim(), 'Process')
        }
    }
}
if (-not $env:MAX_BOT_TOKEN) { throw 'MAX_BOT_TOKEN is missing from .env' }
if (Get-NetTCPConnection -LocalPort 8084 -State Listen -ErrorAction SilentlyContinue) { throw 'Port 8084 is already in use. Stop the existing live bot before starting another poller.' }
$env:DEMO_MODE = 'false'
$env:BOT_MODE = 'polling'
$env:DATA_DIR = 'data/live'
$env:LISTEN_ADDR = '127.0.0.1:8084'
Write-Host 'Live MAX bot: polling. Teams and accounts are stored in data/live.'
Write-Host 'Stop with Ctrl+C. Do not run a second polling process for this bot.'
go run ./cmd/docspace
