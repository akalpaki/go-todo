run:
	@docker compose -f compose.yaml up

test:
	@docker compose -f test_compose.yaml up --abort-on-container-exit -V --force-recreate --build
	@docker compose -f test_compose.yaml down --rmi local -v
	@docker image prune -f
	@docker volume prune -f