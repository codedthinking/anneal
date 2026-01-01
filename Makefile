.PHONY: build run clean test lint

GO := /opt/homebrew/bin/go

build:
	$(GO) build -o bin/tuimail ./main.go

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
	cp bin/tuimail /usr/local/bin/tuimail
