test:
  go test work

clean:
  go clean -testcache

cleanall:
  go clean -cache

up:
  docker compose up -d

up-infra:
  docker compose up rabbitmq db opensearch data-prepper otel-collector opensearch-dashboards -d

down:
  docker compose down

[parallel]
run: run-publisher run-consumer

[working-directory: 'services/publisher/cmd/publisher']
run-publisher:
  go run main.go

[working-directory: 'services/consumer/cmd/consumer']
run-consumer:
  go run main.go

[parallel]
debug: debug-publisher debug-consumer

[working-directory: 'services/publisher/cmd/publisher']
debug-publisher:
  go run -gcflags="all=-N -l" main.go

[working-directory: 'services/consumer/cmd/consumer']
debug-consumer:
  go run -gcflags="all=-N -l" main.go
