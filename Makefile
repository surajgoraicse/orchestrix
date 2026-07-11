
.PHONY: docker-up, docker-down

docker-up:
	docker compose -f ./infra/docker-compose.yaml --env-file ./services/scheduler/.env up -d


docker-down:
	docker compose -f ./infra/docker-compose.yaml --env-file ./services/scheduler/.env down


