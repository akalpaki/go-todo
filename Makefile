build:
	@go build -o bin/ ./...

run: build
	@./bin/todo

test:
	@go test -race -shuffle=on ./... 

docker_run:
	@docker compose --env-file .env.dev up

docker_clean:
	@docker compose rm -f
	@docker image rm todo-todo-server