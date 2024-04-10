build:
	@go build -o bin/ ./...

run: build
	@./bin/todo

test:
	@go test -race -shuffle=on ./... 