.DEFAULT_GOAL := help

.PHONY: serve build clean generate fonts help

generate: ## Fetch spec content and generate Hugo pages
	go run ./cmd/oasis-site-build/

fonts: ## Download self-hosted web fonts (idempotent)
	bash scripts/fetch-fonts.sh

serve: generate fonts ## Generate content, fetch fonts, then start local development server with drafts enabled
	hugo server -D

build: generate fonts ## Generate content, fetch fonts, then build the site with minification
	hugo --minify

clean: ## Remove build output, generated content, and resource cache
	rm -rf public/ resources/ content/en/docs/ data/versions.yaml .cache/

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'
