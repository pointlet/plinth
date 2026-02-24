SQLC := go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0

SQLITE_DIR := internal/core/storage/metadata/sqlite
SQLC_CONFIG := $(SQLITE_DIR)/sqlc.yaml
MIGRATIONS_DIR := $(SQLITE_DIR)/migrations
DB ?= /tmp/plinth.db
GOCACHE ?= /tmp/plinth-gocache

.PHONY: sqlc-generate sqlc-check db-up db-status test ci

sqlc-generate:
	$(SQLC) generate -f $(SQLC_CONFIG)

sqlc-check:
	@test -z "$$(git status --porcelain -- $(SQLITE_DIR))" || (git status --short -- $(SQLITE_DIR); echo "sqlite metadata tree is dirty after generation"; exit 1)

db-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) sqlite3 $(DB) up

db-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) sqlite3 $(DB) status

test:
	GOCACHE=$(GOCACHE) go test ./...

ci: sqlc-generate sqlc-check test
