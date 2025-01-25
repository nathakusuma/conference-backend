# Makefile

# Use the .env file to load environment variables
-include .env

# Default environment variables for Docker Compose
POSTGRES_URL := postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Run migration commands
MIGRATE_CMD=docker compose run --rm migrate -database "${POSTGRES_URL}" -path migration/

# Targets for different migration commands
.PHONY: up down redo status version force

# Apply all migrations
migrate-up:
	$(MIGRATE_CMD) up

# Rollback the most recent migration
migrate-down:
	$(MIGRATE_CMD) down

# Revert the most recent migration and then reapply it
migrate-redo:
	$(MIGRATE_CMD) redo

# Show the status of migrations
migrate-status:
	$(MIGRATE_CMD) status

# Show the current version of the database
migrate-version:
	$(MIGRATE_CMD) version

# Force a specific version (replace <version> with the desired version)
migrate-force:
	$(MIGRATE_CMD) force $(version)
