lint:
	golangci-lint run --fix

test:
	go tool gotestsum
