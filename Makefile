up:
	docker compose up -d

down:
	docker compose down

reset-db:
	docker compose down -v
	rm -rf ./data/postgres
	docker compose up -d