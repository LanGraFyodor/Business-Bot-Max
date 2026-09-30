param([switch]$Demo, [switch]$SkipBuild)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
if (Test-Path -LiteralPath '.env') {
    foreach ($line in Get-Content -LiteralPath '.env') {
        if ($line -match '^\s*([A-Z_][A-Z0-9_]*)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2].Trim(), 'Process')
        }
    }
}
if ($Demo -or -not (Test-Path -LiteralPath '.env')) {
    $env:DEMO_MODE = 'true'
    $env:BOT_MODE = 'off'
}
if (-not $SkipBuild) {
    Push-Location web
    try {
        npm.cmd ci --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) { throw 'npm ci failed' }
        npm.cmd run build
        if ($LASTEXITCODE -ne 0) { throw 'React build failed' }
    } finally { Pop-Location }
}
go run ./cmd/docspace
