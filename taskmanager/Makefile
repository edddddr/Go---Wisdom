DATABASE_URL ?= postgres://teamtasks:teamtasks@localhost:5434/teamtasks?sslmode=disable

.PHONY: db-up migrate-up migrate-down run test

db-up:
	docker compose up -d db

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

test:
	go test -race ./...