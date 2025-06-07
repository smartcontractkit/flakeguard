lint:
	golangci-lint run --fix

test:
	go tool gotestsum

examples:
	go run . -c -- -- ./example_tests/... -tags examples
