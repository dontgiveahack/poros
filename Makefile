BINARY := bin/poros
COMPOSE ?= $(shell \
	if podman compose version >/dev/null 2>&1; then echo "podman compose"; \
	elif command -v podman-compose >/dev/null 2>&1; then echo "podman-compose"; \
	else echo "docker compose"; fi)

.PHONY: build test fmt vet lint run clean

build:
	go build -o $(BINARY) ./cmd/poros

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

run: build
	./$(BINARY)

clean:
	rm -rf bin

# web frontend

web-install:
	cd web && bun install

web-dev:
	cd web && bun run dev

web-build:
	cd web && bun run build

web-test:
	cd web && bun run test

# database
db-up:
	$(COMPOSE) up -d --wait

db-down:
	$(COMPOSE) down

db-test:
	POROS_DB="postgres://poros:poros@localhost:5432/poros?sslmode=disable" go test ./internal/pgstre/ -v -count=1

db-logs:
	$(COMPOSE) logs -f db

db-psql:
	$(COMPOSE) exec db psql -U poros -c "SELECT id, data->>'title' FROM transactions ORDER BY id;"

# docker
docker-build:
	podman build -t poros:dev .

docker-up:
	podman compose up --build -d

docker-down:
	podman compose down
