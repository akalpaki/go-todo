build:
	@go build -o bin/todo

run: build
	@./bin/todo

test:
	@go test -race -shuffle=on ./... 