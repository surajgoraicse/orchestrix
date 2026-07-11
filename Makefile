
.PHONY: docker-up docker-down build-scheduler run-scheduler

docker-up:
	docker compose -f ./infra/docker-compose.yaml --env-file ./services/scheduler/.env up -d


docker-down:
	docker compose -f ./infra/docker-compose.yaml --env-file ./services/scheduler/.env down



# services

build-scheduler:
	@echo "building the scheduler"
	docker build -f services/scheduler/Dockerfile -t scheduler-service:latest .

# run-scheduler:
# 	@echo "building the scheduler"
# 	$(MAKE) build-scheduler
# 	@echo "Starting scheduler service..."
# 	docker run --rm --name scheduler-service --network orchestrix --env-file ./services/scheduler/.env scheduler-service:latest

run-scheduler:
	docker compose -f ./infra/docker-compose.yaml --env-file ./services/scheduler/.env up scheduler