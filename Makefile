.PHONY: build run test migrate

build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test ./...

migrate:
	go run cmd/migrate/main.go

.DEFAULT_GOAL := run
