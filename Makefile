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
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -cover ./... -run TestIntegration
	@$(MAKE) test_coverage_report

test_full: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	@mkdir -p $(GOCOVERDIR)/unit
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -coverprofile=./coverage/unit.out ./... -args -test.gocoverdir=$(GOCOVERDIR)/unit
	@$(MAKE) test_coverage_report

test_full_race: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	@mkdir -p $(GOCOVERDIR)/unit
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -coverprofile=./coverage/unit.out -race ./... -args -test.gocoverdir=$(GOCOVERDIR)/unit
	@$(MAKE) test_coverage_report

# Generate coverage reports from collected data
test_coverage_report:
	@echo "Code coverage"
	@echo "--------------------------------"
	@if [ -d "coverage/unit" ]; then \
		go tool covdata textfmt -i=coverage/unit -o=coverage/unit.out; \
		go tool cover -html=coverage/unit.out -o=coverage/unit.html; \
		echo "Unit"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/unit; \
	fi
	@if [ -d "coverage/integration" ]; then \
		go tool covdata textfmt -i=coverage/integration -o=coverage/integration.out; \
		go tool cover -html=coverage/integration.out -o=coverage/integration.html; \
		echo "--------------------------------"; \
		echo "Integration"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/integration; \
	fi
	@if [ -d "coverage/unit" ] && [ -d "coverage/integration" ]; then \
		mkdir -p coverage/combined; \
		go tool covdata merge -i=coverage/unit,coverage/integration -o coverage/combined; \
		go tool covdata textfmt -i=coverage/combined -o=coverage/combined.out; \
		go tool cover -html=coverage/combined.out -o=coverage/combined.html; \
		echo "--------------------------------"; \
		echo "Combined"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/combined; \
	fi

# Clean coverage data
clean_coverage:
	@rm -rf coverage/
