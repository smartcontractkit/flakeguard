.PHONY: build lint test test_verbose test_unit test_unit_verbose test_race test_full test_integration test_coverage_report clean_coverage

build:
	@goreleaser check
	goreleaser build --snapshot --single-target --clean

lint:
	golangci-lint run --fix

test:
	go tool gotestsum -- -cover ./...

test_race:
	go tool gotestsum -- -cover -race ./...

# Set default coverage directory if not provided
GOCOVERDIR ?= $(PWD)/coverage

test_unit:
	go tool gotestsum -- -cover -short ./...

test_integration: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	-GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -cover ./... -run TestIntegration
	@$(MAKE) test_coverage_report

test_full: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	-GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -cover -coverprofile=./coverage/unit.out -covermode=atomic ./...
	@$(MAKE) test_coverage_report

test_full_race: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	-GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -cover -coverprofile=./coverage/unit.out -covermode=atomic -race ./...
	@$(MAKE) test_coverage_report

# Generate coverage reports from collected data
test_coverage_report:
	@echo "Code coverage"
	@echo "--------------------------------"
	@if [ -f "coverage/unit.out" ]; then \
		go tool cover -html=coverage/unit.out -o=coverage/unit.html; \
		echo "Unit tests:"; \
		go tool cover -func=coverage/unit.out | tail -1; \
	fi
	@if [ -d "coverage/integration" ]; then \
		go tool covdata textfmt -i=coverage/integration -o=coverage/integration.out; \
		go tool cover -html=coverage/integration.out -o=coverage/integration.html; \
		echo "Integration tests:"; \
		go tool covdata percent -i=coverage/integration; \
	fi
	@if [ -f "coverage/unit.out" ] && [ -f "coverage/integration.out" ]; then \
		go tool covdata textfmt -i=coverage/integration -o=coverage/integration.out; \
		go run github.com/wadey/gocovmerge coverage/unit.out coverage/integration.out > coverage/combined.out; \
		go tool cover -html=coverage/combined.out -o=coverage/combined.html; \
		echo "Combined coverage:"; \
		go tool cover -func=coverage/combined.out | tail -1; \
	fi

# Clean coverage data
clean_coverage:
	@rm -rf coverage/
