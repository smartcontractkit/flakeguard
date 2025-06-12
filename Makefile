.PHONY: build lint test test_verbose test_unit test_unit_verbose test_race

build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_verbose:
	FLAKEGUARD_TEST_LOG_LEVEL=debug go tool gotestsum -- -cover -v ./...

test_unit:
	go tool gotestsum -- -cover ./... -short

test_unit_verbose:
	FLAKEGUARD_TEST_LOG_LEVEL=debug go tool gotestsum -- -cover -v -short ./...

test_race:
	go tool gotestsum -- -cover -race ./...
