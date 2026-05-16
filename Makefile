COMPOSE_RUN = docker compose run --rm dev

.PHONY: test build tidy shell
test:
	$(COMPOSE_RUN) go test ./...

build:
	$(COMPOSE_RUN) go build -o sage .

tidy:
	$(COMPOSE_RUN) go mod tidy

shell:
	$(COMPOSE_RUN) bash
