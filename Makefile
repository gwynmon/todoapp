ifneq (,$(wildcard .env))
    include .env
    export
endif

POSTGRES_DSN ?= postgres://todouser:todopassword@localhost:5432/tododb?sslmode=disable

.PHONY: init migrate-up migrate-down migrate-create run up down logs test

init:
	@[ -f .env ] || cp .env.example .env

migrate-up:
	@goose -dir ./migrations postgres "$(POSTGRES_DSN)" up

migrate-down:
	@goose -dir ./migrations postgres "$(POSTGRES_DSN)" down

migrate-create:
	@goose -dir ./migrations create $(name) sql

run:
	@go run ./cmd/app

up:
	@docker compose up --build -d

down:
	@docker compose down

logs:
	@docker compose logs -f app

test:
	@go test ./...