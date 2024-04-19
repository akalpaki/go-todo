run:
	@docker compose -f compose.yaml up

test:
	@docker compose -f test_compose.yaml up --abort-on-container-exit -V --force-recreate