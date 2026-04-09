.DEFAULT_GOAL := help

.PHONY: serve build clean help

serve: ## Start local development server with drafts enabled
	hugo server -D

build: ## Build the site with minification
	hugo --minify

clean: ## Remove build output and resource cache
	rm -rf public/ resources/

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
