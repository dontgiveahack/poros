BINARY := bin/poros

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
