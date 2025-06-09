build:
	go build -o flakeguard ./cmd/flakeguard/main.go

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_race:
	go tool gotestsum -- -cover -race ./...

detect_examples_flaky:
	go run ./cmd/flakeguard/main.go detect -c -L debug -- -- ./example_tests/flaky/... -tags examples

guard_examples_flaky:
	go run ./cmd/flakeguard/main.go guard -c -L debug -- -- ./example_tests/flaky/... -tags examples

detect_examples_race:
	go run ./cmd/flakeguard/main.go detect -c -L debug -- -- ./example_tests/race/... -race -tags examples

guard_examples_race:
	go run ./cmd/flakeguard/main.go guard -c -L debug -- -- ./example_tests/race/... -race -tags examples

detect_examples_panic:
	go run ./cmd/flakeguard/main.go detect -c -L debug -- -- ./example_tests/panic/... -tags examples

guard_examples_panic:
	go run ./cmd/flakeguard/main.go guard -c -L debug -- -- ./example_tests/panic/... -tags examples
