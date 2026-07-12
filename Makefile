
.PHONY: docker-up docker-down build-scheduler run-scheduler protoc-coordinator

docker-up:
	docker compose -f infra/docker-compose.yaml up -d postgres 


docker-down:
	docker compose -f infra/docker-compose.yaml rm -fs postgres 



# proto

protoc-coordinator:
	@echo "generating proto for coordinator..."
	@mkdir -p api/gen/go/coordinator/v1
	@protoc \
		-I api/proto/coordinator/v1 \
		--go_out=api/gen/go/coordinator/v1 --go_opt=paths=source_relative \
		--go-grpc_out=api/gen/go/coordinator/v1 --go-grpc_opt=paths=source_relative \
		coordinator.proto
	@echo "successfully generated proto for coordinator"

# generate proto for worker
protoc-worker:
	@echo "generating proto for worker..."
	@mkdir -p api/gen/go/worker/v1
	@protoc \
		-I api/proto/worker/v1 \
		--go_out=api/gen/go/worker/v1 --go_opt=paths=source_relative \
		--go-grpc_out=api/gen/go/worker/v1 --go-grpc_opt=paths=source_relative \
		worker.proto
	@echo "successfully generated proto for worker"


# services
build-scheduler:
	@echo "building the scheduler"
	docker build -f services/scheduler/Dockerfile -t scheduler-service:latest .

run-scheduler:
	$(MAKE) build-scheduler
	docker compose -f infra/docker-compose.yaml up scheduler

stop-scheduler:
	docker compose -f infra/docker-compose.yaml rm -fs scheduler