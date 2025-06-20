.PHONY: build lint test test_verbose test_unit test_unit_verbose test_race test_full test_integration test_coverage_report clean_coverage

build:
	@goreleaser check
	goreleaser build --snapshot --single-target --clean

lint:
	golangci-lint run --fix

bench:
	FLAKEGUARD_TEST_LOG_LEVEL=disabled go test -bench=. -benchmem -run=^$$ ./... -cpu=2,4,8

test:
	go tool gotestsum -- -cover ./...

test_race:
	go tool gotestsum -- -cover -race ./...

# Set default coverage directory if not provided
GOCOVERDIR ?= $(PWD)/coverage

test_short:
	go tool gotestsum -- -cover -short ./...

test_unit:
	go tool gotestsum -- -cover -skip TestIntegration ./...

test_integration: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- ./... -run TestIntegration
	@$(MAKE) test_coverage_report

# -args -test.gocoverdir=$(GOCOVERDIR)/unit is using an experimental flag to use the newer go coverage format for unit tests along with integration tests.
# Read more at the official code here: https://github.com/golang/go/issues/51430#issuecomment-1344711300
# And this handy blog post: https://dustinspecker.com/posts/go-combined-unit-integration-code-coverage/

test_full: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	@mkdir -p $(GOCOVERDIR)/unit
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -cover -covermode=atomic ./... -args -test.gocoverdir=$(GOCOVERDIR)/unit
	@$(MAKE) test_coverage_report

test_full_race: clean_coverage build
	@mkdir -p $(GOCOVERDIR)
	@mkdir -p $(GOCOVERDIR)/unit
	-FLAKEGUARD_GOCOVERDIR=$(GOCOVERDIR) go tool gotestsum -- -count=1 -cover -covermode=atomic -race ./... -args -test.gocoverdir=$(GOCOVERDIR)/unit
	@$(MAKE) test_coverage_report

# Generate coverage reports from collected data
test_coverage_report:
	@if [ -d "coverage/unit" ]; then \
		go tool covdata textfmt -i=coverage/unit -o=coverage/unit.out; \
		go tool cover -html=coverage/unit.out -o=coverage/unit.html; \
		echo "--------------------------------"; \
		echo "Unit: coverage/unit.html"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/unit; \
		echo ""; \
		printf "🎯 Unit Coverage: "; \
		go tool cover -func=coverage/unit.out | tail -1 | awk '{print $$NF}'; \
	fi
	@if [ -d "coverage/integration" ]; then \
		go tool covdata textfmt -i=coverage/integration -o=coverage/integration.out; \
		go tool cover -html=coverage/integration.out -o=coverage/integration.html; \
		echo "--------------------------------"; \
		echo "Integration: coverage/integration.html"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/integration; \
		echo ""; \
		printf "🎯 Integration Coverage: "; \
		go tool cover -func=coverage/integration.out | tail -1 | awk '{print $$NF}'; \
	fi


	@if [ -d "coverage/unit" ] && [ -d "coverage/integration" ]; then \
		mkdir -p coverage/combined; \
		go tool covdata merge -i=coverage/unit,coverage/integration -o coverage/combined; \
		go tool covdata textfmt -i=coverage/combined -o=coverage/combined.out; \
		sed -i '' 's|$(PWD)/|github.com/smartcontractkit/flakeguard/|g' coverage/combined.out; \
		sed -i '' 's|$(PWD)/|github.com/smartcontractkit/flakeguard/|g' coverage/integration.out; \
		go tool cover -html=coverage/combined.out -o=coverage/combined.html; \
		echo "--------------------------------"; \
		echo "Combined: coverage/combined.html"; \
		echo "--------------------------------"; \
		go tool covdata percent -i=coverage/combined; \
		echo ""; \
		printf "🎯 Total Coverage: "; \
		go tool cover -func=coverage/combined.out | tail -1 | awk '{print $$NF}'; \
	fi
# @echo "go-test-coverage"
# @echo "--------------------------------"
# @go tool go-test-coverage --config=./.test-coverage.yaml

# Clean coverage data
clean_coverage:
	@rm -rf coverage/
