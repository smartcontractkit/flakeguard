.PHONY: build lint test test_verbose test_unit test_unit_verbose test_race test_integration

build:
	goreleaser check
	goreleaser build --snapshot --single-target --clean

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_unit:
	go tool gotestsum -- -cover -short ./...

test_race:
	go tool gotestsum -- -cover -race ./...

test_integration:
	go tool gotestsum -- -cover ./... -run TestIntegration
