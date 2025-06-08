build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_race:
	go tool gotestsum -- -cover -race ./...

detect_examples_flaky:
	go run ./cmd/flakeguard/main.go detect -c -L debug -- -- ./example_tests/some_flaky/... -tags examples

guard_examples_flaky:
	go run ./cmd/flakeguard/main.go guard -c -L debug -- -- ./example_tests/some_flaky/... -tags examples

detect_examples_panic:
	go run ./cmd/flakeguard/main.go detect -c -L debug -- -- ./example_tests/some_panic/... -tags examples

guard_examples_panic:
	go run ./cmd/flakeguard/main.go guard -c -L debug -- -- ./example_tests/some_panic/... -tags examples
