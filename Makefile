build:
	@go build -o bin/previewer

start:
	@./bin/previewer

test:
	@go test ./... -v
