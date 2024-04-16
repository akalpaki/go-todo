run:
	@docker compose -f compose.yaml up

test:
	@docker compose -f test_compose.yaml up
	
clean:
	@docker compose down --rmi local
	@docker compose rm -f todo-db-test