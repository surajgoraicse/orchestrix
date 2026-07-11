# --- Build Stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# 1. Copy the Go workspace files
COPY go.work go.work.sum ./

# 2. Copy the shared libraries
COPY libs/ libs/

# 3. Copy the specific service we want to build
COPY services/scheduler/ services/scheduler/

# 4. Change directory to the service and build the binary
WORKDIR /workspace/services/scheduler
ENV GOWORK=off
RUN go build -o /bin/scheduler main.go

# --- Final Stage ---
FROM alpine:latest
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /bin/scheduler /app/scheduler

# COPY --from=builder /workspace/services/scheduler/db/migrations ./migrations
COPY --from=builder /workspace/services/scheduler/.env ./.env

EXPOSE 8000

CMD ["/app/scheduler"]