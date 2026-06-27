APP_NAME := links-helper-bot

.PHONY: test vet build up down logs compose-config

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o bin/$(APP_NAME) .

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f bot

compose-config:
	docker compose config
