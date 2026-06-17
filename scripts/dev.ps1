param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("migrate", "seed", "sqlc", "test", "frontend-build", "user-dev", "admin-dev")]
  [string]$Task
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

switch ($Task) {
  "migrate" {
    Push-Location "$Root\backend"
    go run ./cmd/migrate
    Pop-Location
  }
  "seed" {
    Push-Location "$Root\backend"
    go run ./cmd/seed
    Pop-Location
  }
  "sqlc" {
    Push-Location "$Root\backend"
    go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
    Pop-Location
  }
  "test" {
    Push-Location "$Root\backend"
    go test ./...
    Pop-Location
  }
  "frontend-build" {
    Push-Location "$Root\frontend\user"
    npm run build
    Pop-Location
    Push-Location "$Root\frontend\admin"
    npm run build
    Pop-Location
  }
  "user-dev" {
    Push-Location "$Root\frontend\user"
    npm run dev
    Pop-Location
  }
  "admin-dev" {
    Push-Location "$Root\frontend\admin"
    npm run dev
    Pop-Location
  }
}
