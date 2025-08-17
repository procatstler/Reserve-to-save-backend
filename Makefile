up:
	docker compose up -d

down:
	docker compose down

reset-db:
	docker compose down -v
	rm -rf ./pkg/db/data/*
	docker compose up -d

build:
	go build -o api-server/api-server ./api-server/main.go
	go build -o query-server/query-server ./query-server/main.go

start:
	make up
	./api-server/api-server &
	./query-server/query-server &

stop:
	killall api-server
	killall query-server
	make down