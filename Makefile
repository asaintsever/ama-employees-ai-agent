.SILENT: ;               # No need for @
.ONESHELL: ;             # Single shell for a target (required to properly use all of our local variables)
.EXPORT_ALL_VARIABLES: ; # Send all vars to shell
.DEFAULT: help # Running Make without target will run the help target

.PHONY: help build clean test

help: ## Show Help
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	# Check if SLACK_TOKEN is set
	if [ -z "$$SLACK_TOKEN" ]; then \
		echo "⚠️  Warning: SLACK_TOKEN environment variable is not set"; \
		echo "   Setting a test token (will cause Slack authentication errors)"; \
		export SLACK_TOKEN="xoxp-test-token"; \
	fi
	# Check if AWS credentials are configured
	if [ -z "$$AWS_ACCESS_KEY_ID" ] || [ -z "$$AWS_SECRET_ACCESS_KEY" ]; then \
		echo "⚠️  Warning: AWS credentials are not set"; \
		echo "   Tests will fail with authentication errors (expected)"; \
	fi
	go test -v ./pkg/agent

clean: ## Clean build artifacts
	rm -f target/*

build: clean ## Build binary
	set -e
	go build -o target/ama-employees-ai-agent ./cmd/agent