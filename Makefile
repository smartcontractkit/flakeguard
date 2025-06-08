build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_race:
	go tool gotestsum -- -cover -race ./...

test_examples:
	go run ./cmd/flakeguard/main.go -c -- -- ./example_tests/... -tags examples
