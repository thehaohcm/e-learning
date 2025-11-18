.PHONY: up down logs api web db

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

api:
	docker compose exec api sh -c "./server"

db:
	docker compose exec db psql -U $$DB_USER -d $$DB_NAME

clean:
	docker compose down -v
