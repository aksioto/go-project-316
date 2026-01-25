.PHONY: tidy fmt test test-race cover lint build run check

tidy:
	go mod tidy
	go mod verify

fmt:
	goimports -w .

test:
	go test -v ./...

test-race:
	go test -race ./...

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out
	rm cover.out

lint:
	golangci-lint run ./...

build:
	go build -o bin/hexlet-go-crawler ./cmd/hexlet-go-crawler

run:
	@if [ -z "$(URL)" ]; then \
		go run ./cmd/hexlet-go-crawler --help; \
	else \
		go run ./cmd/hexlet-go-crawler $(URL); \
	fi

check: tidy fmt lint test