SERVICES := training auth analytics notification

.PHONY: help init up down build logs migrate-training

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

# ── Database ───────────────────────────────────────────────────────────────────

migrate-training: ## Apply training service migration (requires running postgres)
	docker compose exec -T postgres psql -U postgres -d training_db \
		< services/training/migrations/001_create_training_sessions.sql

psql: ## Open psql shell in training_db
	docker compose exec postgres psql -U postgres -d training_db
