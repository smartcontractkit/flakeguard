build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_unit:
	go tool gotestsum -- -cover ./... -short

test_race:
	go tool gotestsum -- -cover -race ./...

test_examples:
	go tool gotestsum -- -cover TestExample
