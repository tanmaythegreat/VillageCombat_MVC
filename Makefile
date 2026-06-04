
# Your defaults are preserved here, but you can override them via terminal if needed
DB_USER ?= admin_tentellam
DB_PASS ?= i_wont_tell_you
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= VillageGameDB
DB_SSL  ?= disable

DB_URL = postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)
MIGRATIONS_PATH = db/migrations

.PHONY: migrate-current migrate-up migrate-down migrate-create db-force-clear help

## migrate-current: Check which migration version the database is currently running
migrate-current:
	@echo "Checking current active database schema version..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

## migrate-up: Push all pending changes UP to the database
migrate-up:
	@echo "Applying all pending database migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

## migrate-down: Roll back (undo) the single last applied migration step
migrate-down:
	@echo "Rolling back the last migration step..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
## migrate-create name=your_table_name: Generate new sequential migration files
migrate-create:
ifndef name
	$(error ERROR: You must specify a migration name. Example: make migrate-create name=add_users_table)
endif
	@echo "Creating sequential migration template files for: $(name)"
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

## db-force-clear version=X: Force-reset a dirty/broken migration ledger back to a clean state
db-force-clear:
ifndef version
	$(error ERROR: You must specify the version to force mark. Example: make db-force-clear version=1)
endif
	@echo "Force-setting migration state ledger to version: $(version)"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)
## help: Show a summary of all available administrative tools
help:
	@echo "Available administrative database commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'