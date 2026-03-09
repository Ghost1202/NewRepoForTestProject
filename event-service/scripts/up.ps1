$ErrorActionPreference = "Stop"

$connectUrl      = "http://127.0.0.1:8083"
$connectorName   = "event-source"
$configPath      = Join-Path $PSScriptRoot "..\event-source-config.json"

function Require-File([string]$path) {
  if (-not (Test-Path $path)) {
    throw "File not found: $path"
  }
}

function Wait-Http200([string]$url, [int]$timeoutSec = 90) {
  $deadline = (Get-Date).AddSeconds($timeoutSec)
  while ((Get-Date) -lt $deadline) {
    try {
      $r = Invoke-WebRequest -UseBasicParsing -Uri $url -Method GET -TimeoutSec 5
      if ($r.StatusCode -eq 200) { return }
    } catch { }
    Start-Sleep -Seconds 2
  }
  throw "Timeout waiting for HTTP 200: $url"
}

function Wait-ComposeHealthy([int]$timeoutSec = 120) {
  $deadline = (Get-Date).AddSeconds($timeoutSec)
  while ((Get-Date) -lt $deadline) {
    $lines = docker compose ps --format json 2>$null
    if (-not $lines) { Start-Sleep 2; continue }

    $items = $lines | ForEach-Object { $_ | ConvertFrom-Json }
    $bad = @()

    foreach ($it in $items) {
      $state = ($it.State  | Out-String).Trim()
      $health = ($it.Health | Out-String).Trim()

      $ok =
        ($state -match "running") -and (
          ($health -eq "") -or
          ($health -match "healthy")
        )

      if (-not $ok) { $bad += "$($it.Service): state=$state health=$health" }
    }

    if ($bad.Count -eq 0) { return }

    Start-Sleep -Seconds 2
  }

  docker compose ps
  throw "Timeout waiting containers to become Running/Healthy"
}

Write-Host "==> docker compose up -d"
docker compose up -d

Write-Host "==> wait compose Running/Healthy"
Wait-ComposeHealthy 180

Write-Host "==> wait Kafka Connect is alive: $connectUrl/"
Wait-Http200 "$connectUrl/" 120

Require-File $configPath

Write-Host "==> upsert connector config (PUT /connectors/$connectorName/config)"
curl.exe -sS -i -X PUT "$connectUrl/connectors/$connectorName/config" `
  -H "Content-Type: application/json" `
  --data-binary "@$configPath" | Out-Host

Write-Host "==> connector status"
Invoke-RestMethod "$connectUrl/connectors/$connectorName/status" | ConvertTo-Json -Depth 20 | Out-Host

Write-Host "==> DONE"

//powershell -ExecutionPolicy Bypass -File .\scripts\up.ps1
