SERVICES := training auth analytics notification

.PHONY: help init up down build logs training-logs \
        migrate-up migrate-down migrate-version migrate-create psql

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Setup ──────────────────────────────────────────────────────────────────────

init: ## Run go mod tidy for all services (run this first)
	@for svc in $(SERVICES); do \
		echo "→ go mod tidy: services/$$svc"; \
		cd services/$$svc && go mod tidy && cd ../..; \
	done

# ── Docker Compose ─────────────────────────────────────────────────────────────

up: ## Start all active services
	docker compose up -d

down: ## Stop all services
	docker compose down

build: ## Rebuild all service images
	docker compose build

logs: ## Tail logs from all active services
	docker compose logs -f

# ── Per-service logs ───────────────────────────────────────────────────────────

training-logs: ## Tail training service logs
	docker compose logs -f training-service

# ── Migrations ─────────────────────────────────────────────────────────────────

migrate-up: ## Apply all pending training migrations
	docker compose run --rm migrate-training

migrate-down: ## Roll back the last training migration
	docker compose run --rm migrate-training \
		-path=/migrations \
		-database="postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@postgres:5432/training_db?sslmode=disable" \
		down 1

migrate-version: ## Show current training migration version
	docker compose run --rm migrate-training \
		-path=/migrations \
		-database="postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@postgres:5432/training_db?sslmode=disable" \
		version

migrate-create: ## Create a new migration pair: make migrate-create name=add_something
	@[ -n "$(name)" ] || (echo "Usage: make migrate-create name=<description>"; exit 1)
	@seq=$$(ls services/training/migrations/*.up.sql 2>/dev/null | wc -l | tr -d ' '); \
	 seq=$$(printf "%06d" $$((seq + 1))); \
	 touch services/training/migrations/$${seq}_$(name).up.sql; \
	 touch services/training/migrations/$${seq}_$(name).down.sql; \
	 echo "Created: $${seq}_$(name).up.sql / .down.sql"

# ── Database ───────────────────────────────────────────────────────────────────

psql: ## Open psql shell in training_db
	docker compose exec postgres psql -U postgres -d training_db
