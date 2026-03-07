test:
  go test work

clean:
  go clean -testcache

cleanall:
  go clean -cache

up:
  docker compose up -d

down:
  docker compose down
