up:
	docker compose up -d

down:
	docker compose down

reset-db:
	docker compose down -v
	rm -rf ./pkg/db/data/*
	docker compose up -d