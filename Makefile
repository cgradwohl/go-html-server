build:
	@go build -o bin/go-html-server -v

run: build
	@./bin/go-html-server

test:
	@go test -v ./...
