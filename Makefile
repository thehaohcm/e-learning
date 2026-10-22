.PHONY: up down logs build test api-dev web-dev localdb psql backup

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

build:
	docker compose build $(service)

# Go unit tests
test:
	cd api && go test ./...

# Run the API locally (reads ../.env through your shell, e.g. `set -a; . ./.env; set +a`)
api-dev:
	cd api && go run ./cmd/server

# Run the web app with hot reload; /api is proxied to localhost:8000
web-dev:
	cd web && npm run dev

# Start only PostgreSQL (e.g. for `make api-dev` with DB_HOST=localhost and the db port published)
localdb:
	docker compose up -d db

# Open psql inside the db container (reads DB_USER / DB_NAME from .env)
psql:
	set -a; . ./.env; set +a; docker compose exec db psql -U $$DB_USER -d $$DB_NAME

# Backup the database to ./backups
backup:
	mkdir -p backups; set -a; . ./.env; set +a; docker compose exec -T db pg_dump -U $$DB_USER -Fc $$DB_NAME > backups/db_$$(date +%Y%m%d_%H%M).dump
