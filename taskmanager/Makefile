DATABASE_URL ?= postgres://teamtasks:teamtasks@localhost:5434/teamtasks?sslmode=disable

JWT_SECRET ?= 3f7d1b82e9c40a5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a


.PHONY: db-up migrate-up migrate-down run test

db-up:
	docker compose up -d db

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

run:
	DATABASE_URL="$(DATABASE_URL)" JWT_SECRET="$(JWT_SECRET)" go run ./cmd/api


test:               ## fast: no Docker needed
	go test -race -short ./...

test-all:           ## everything, including DB tests (needs Docker)
	go test -race ./...

cover:
	go test -race -coverprofile=cover.out ./...
	go tool cover -func=cover.out | tail -n 1

cover-html: cover
	go tool cover -html=cover.out