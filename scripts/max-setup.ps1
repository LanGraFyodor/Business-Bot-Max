param([switch]$Subscribe)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
if (Test-Path -LiteralPath '.env') {
    foreach ($line in Get-Content -LiteralPath '.env') {
        if ($line -match '^\s*([A-Z_][A-Z0-9_]*)=(.*)$') { [Environment]::SetEnvironmentVariable($matches[1], $matches[2].Trim(), 'Process') }
    }
}
if (-not $env:MAX_BOT_TOKEN) { throw 'Set MAX_BOT_TOKEN in .env' }
$api = 'https://platform-api2.max.ru'
$headers = @{ Authorization = $env:MAX_BOT_TOKEN }
go run ./cmd/docspace --check-bot
if ($LASTEXITCODE -ne 0) { throw "MAX connection failed" }
if ($Subscribe) {
    if (-not $env:PUBLIC_APP_URL -or -not $env:PUBLIC_APP_URL.StartsWith('https://') -or -not $env:MAX_WEBHOOK_SECRET) { throw 'Set HTTPS PUBLIC_APP_URL and MAX_WEBHOOK_SECRET first' }
    $body = @{ url = $env:PUBLIC_APP_URL.TrimEnd('/') + '/api/max/webhook'; update_types = @('bot_started', 'message_created'); secret = $env:MAX_WEBHOOK_SECRET } | ConvertTo-Json
    Invoke-RestMethod -Method Post -Uri "$api/subscriptions" -Headers $headers -ContentType 'application/json' -Body $body
}
