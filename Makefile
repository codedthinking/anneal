.PHONY: build run clean test lint

GO := /opt/homebrew/bin/go

build:
	$(GO) build -o bin/anneal ./main.go

run:
	$(GO) run ./main.go

clean:
	rm -rf bin/

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

deps:
	$(GO) mod tidy

install: build
	cp bin/anneal /usr/local/bin/anneal
