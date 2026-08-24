BINARY := bin/poros

.PHONY: build test fmt vet run clean

build:
	go build -o $(BINARY) ./cmd/poros

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: build
	./$(BINARY)

clean:
	rm -rf bin
