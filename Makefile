up:
	docker compose up -d

down:
	docker compose down

reset-db:
	docker compose down -v
	rm -rf ./pkg/db/data/*
	docker compose up -d

start:
	make up
	./api-server/api-server &
	./query-server/query-server &

stop:
	killall api-server
	killall query-server
	make down