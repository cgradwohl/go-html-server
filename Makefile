build:
	@go build -o bin/go-html-server -v

run: build
	@./bin/go-html-server

liverun:
	@ls *.go | entr -r make run

test:
	@go test -v ./...
