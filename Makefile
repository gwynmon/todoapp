DB_DSN ?= postgres://todouser:changeme@localhost:5433/tododb?sslmode=disable

init:
	@[ -f .env ] || cp .env.example .env

migrate-up:
	@goose -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	@goose -dir migrations postgres "$(DB_DSN)" down