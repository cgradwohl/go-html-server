build:
	@go build -o bin/go-html-server -v

run: build
	@./bin/go-html-server

liverun:
	@command -v entr >/dev/null 2>&1 || { echo >&2 "entr not found. Install it here: https://github.com/eradman/entr or use 'make run' for manual reloading."; exit 1; }
	@ls *.go | entr -r make run

test:
	@go test -v ./...
