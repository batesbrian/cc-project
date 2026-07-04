.PHONY: help
help: ## print make targets 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: go-install-air
go-install-air: ## Installs the air build reload system using 'go install'
	go install github.com/air-verse/air@latest

.PHONY: get-install-air
get-install-air: ## Installs the air build reload system using cUrl
	curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

.PHONY: go-install-templ
go-install-templ: ## Installs templ and adds as a local project dependency
	go install github.com/a-h/templ/cmd/templ@latest
	go get github.com/a-h/templ

.PHONY: get-install-tailwindcss
get-install-tailwindcss: ## Installs the tailwindcss cli
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 tailwindcss

.PHONY: tailwind-watch
tailwind-watch: ## compile tailwindcss and watch for changes
	./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css --watch

.PHONY: tailwind-build
tailwind-build: ## one-time compile tailwindcss styles
	./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css

.PHONY: build
build: ## compile tailwindcss and templ files and build the project
	./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css
	templ generate
	go build -o ./tmp ./cmd/main.go

.PHONY: live/templ
live/templ: ## regenerate _templ.go files and reload proxy
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false

.PHONY: live/server
live/server: ## rebuild and rerun go server
	air

.PHONY: live/tailwind
live/tailwind: ## recompile css input on change
	./tailwindcss -i ./static/css/custom.css -o ./static/css/style.css --watch

# .PHONY: live/sync-assets
# live/sync-assets: ## reload proxy browser on css recompile
# 	air --build.cmd "templ generate --notify-proxy" --build.bin "/usr/bin/true" \
# 		--build.delay "100" --build.include_dir "static/css" --build.include_ext "css"

.PHONY: dev
dev: ## launch hot reload dev server for tailwind, templ, and go
	make -j3 live/tailwind live/server live/templ

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch
