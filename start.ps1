param(
  [ValidateSet('local','docker')][string]$Mode = 'local',
  [int]$Port = 8080,
  [string]$DbPath = "$PSScriptRoot\data\mss.db",
  [switch]$NoTidy
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Definition

if ($Mode -eq 'local') {
  $serverDir = Join-Path $root 'server'
  $env:MSS_LISTEN_ADDR = ":$Port"
  $env:MSS_DB_PATH = $DbPath
  $dbDir = Split-Path -Parent $DbPath
  if (!(Test-Path $dbDir)) { New-Item -ItemType Directory -Path $dbDir | Out-Null }
  Push-Location $serverDir
  try {
    if (-not $NoTidy) { go mod tidy }
    go run ./cmd/mss-server
  } finally {
    Pop-Location
  }
}
elseif ($Mode -eq 'docker') {
  Push-Location $root
  try {
    docker compose up --build -d
    Write-Host "mss-server started in Docker. URL: http://localhost:8080/healthz" -ForegroundColor Green
  } finally {
    Pop-Location
  }
}
else {
  Write-Error "Unknown mode. Use -Mode local|docker"
  exit 1
}
