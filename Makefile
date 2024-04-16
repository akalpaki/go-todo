run:
	@docker compose -f compose.yaml up

test:
	@docker compose -f test_compose.yaml up

clean:
	@docker compose rm -f
	@docker image rm todo-todo-server