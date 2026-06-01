ifneq (,$(wildcard .env))
	include .env
	export
endif

POSTGRES_DSN ?= postgres://todouser:todopassword@localhost:5432/tododb?sslmode=disable

.PHONY: init migrate-up migrate-down migrate-create run up down

init:
	@[ -f .env ] || cp .env.example .env

migrate-up:
	@goose -dir ./migrations postgres "$(POSTGRES_DSN)" up

migrate-up-docker:
	@docker compose exec -T app sh -c 'go install github.com/pressly/goose/v3/cmd/goose@latest && goose -dir /app/migrations postgres "postgres://todouser:todopassword@postgres:5432/tododb?sslmode=disable" up'

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