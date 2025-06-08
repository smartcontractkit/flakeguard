build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum

test_race:
	go tool gotestsum -- -race ./...

test_examples:
	go run ./cmd/flakeguard/main.go -c -- -- ./example_tests/... -tags examples
