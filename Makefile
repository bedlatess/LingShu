SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command

.PHONY: up down migrate seed sqlc test backend-test frontend-install frontend-build

up:
	docker compose up --build

down:
	docker compose down

migrate:
	cd backend; go run ./cmd/migrate

seed:
	cd backend; go run ./cmd/seed

sqlc:
	cd backend; go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate

test:
	cd backend; go test ./...

backend-test:
	cd backend; go test ./...

frontend-install:
	cd frontend; npm install

frontend-build:
	cd frontend; npm run build
