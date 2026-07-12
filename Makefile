
.PHONY: docker-up docker-down build-scheduler run-scheduler

docker-up:
	docker compose -f infra/docker-compose.yaml up -d postgres 


docker-down:
	docker compose -f infra/docker-compose.yaml rm -fs postgres 



# services
build-scheduler:
	@echo "building the scheduler"
	docker build -f services/scheduler/Dockerfile -t scheduler-service:latest .

run-scheduler:
	$(MAKE) build-scheduler
	docker compose -f infra/docker-compose.yaml up scheduler

stop-scheduler:
	docker compose -f infra/docker-compose.yaml rm -fs scheduler