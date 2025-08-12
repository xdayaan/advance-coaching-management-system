# Project configuration
GITHUB_USER = your-username
GITHUB_REPO = acms_v2_backend

# Load environment variables
include .env
export

# Database URL for migrations
DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# =============================================================================
# MIGRATION COMMANDS
# =============================================================================
migrate-create:
	@read -p "Enter migration name: " name; \
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest create -ext sql -dir migrations $$name

migrate-up:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migrations -database "$(DATABASE_URL)" down 1

migrate-reset:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migrations -database "$(DATABASE_URL)" down

migrate-version:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migrations -database "$(DATABASE_URL)" version

migrate-force:
	@read -p "Enter version to force: " version; \
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path migrations -database "$(DATABASE_URL)" force $$version

# =============================================================================
# APPLICATION COMMANDS
# =============================================================================
run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs

deps:
	go mod tidy

# =============================================================================
# GITHUB COMMANDS
# =============================================================================
git-init:
	git init
	git add .
	git commit -m "Initial commit: Go REST API with PostgreSQL and JWT auth"
	git branch -M main

git-remote:
	git remote add origin https://github.com/$(GITHUB_USER)/$(GITHUB_REPO).git

git-push:
	@read -p "Enter commit message: " msg; \
	git add . && git commit -m "$$msg" && git push

git-push-main:
	git push -u origin main

git-pull:
	git pull origin main

git-setup: git-init git-remote git-push-main

# =============================================================================
# QUICK COMMANDS
# =============================================================================
dev: migrate-up run

deploy: deps swagger build

clean:
	rm -rf bin/ docs/ coverage.out

.PHONY: migrate-create migrate-up migrate-down migrate-reset migrate-version migrate-force run build test swagger deps git-init git-remote git-push git-push-main git-pull git-setup dev deploy clean