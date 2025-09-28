# Introduction

This repository is a simple demo playground to test out golang publisher/consumer microservices using RabbitMQ as a message broker and docker with hot reload using air.

## Getting started

Open this repository in your favorite code editor and use `docker compose up` to startup the development stack and start coding. The go applications will be running within docker containers with the code mounted to be automatically detected for hot reload using [Air](https://github.com/air-verse/air)

## Guiding Principles

This repository is a simple showcase to practice and use as a reference. The goal is to learn and follow general good software principles found in various guidelines defined bellow.

- [The Twelve-Factor App](https://12factor.net)
- [Effective Go](https://go.dev/doc/effective_go)

## Architecture Overview

```mermaid
    C4Context
      title System Context diagram
      Boundary(b0, "") {
        System(publisher, "Publisher", "Rest API and background publisher<br>:8080")
        System(rabbitmq, "RabbitMQ", "Message broker")
        System(consumer, "Consumer", "Background worker consuming and processing <br>messages from the message broker")
        System(postgresql, "Postgresql", "Database")
      }

      Rel(publisher, rabbitmq, "Publishes to")
      Rel(consumer, rabbitmq, "Subscribes to")
      Rel(consumer, postgresql, "Read / Write")
      UpdateLayoutConfig($c4ShapeInRow="1", $c4BoundaryInRow="1")

```

## Software Architecture Flow

```mermaid
  flowchart LR
    A.1[http handler] --> B(application service)
    A.2[consumer] --> B(application service)
    B --> C{readonly action <br>or <br>mutate data?}
    C --> D[query]
    C --> E[command]
    D -->|reads| F[repository]
    E -->|writes| F[repository]
    F --> G[database]
```

### Demo concepts

| ID | Name | Status |
| - | - | - |
| 1 | Code live reload | x |
| 2 | Dev docker containers | x |
| 3 | Prod docker containers | |
| 4 | Application configuration | x |
| 5 | Http Endpoint | x |
| 6 | Http Endpoint Versioning | x |
| 7 | Middleware | x |
| 8 | Dependency Injection | x |
| 9 | Structured Logging | x |
| 10 | OpenTelemetry Instrumentation | |
| 11 | Message Broker Publisher | x |
| 12 | Message Broker Consumer | x |
| 13 | Message Broker Outbox Pattern | |
| 14 | Message Broker Inbox Pattern | |
| 15 | Database query | x |
| 16 | Database updates | x |
| 17 | Database transactions | |
| 18 | Database migrations | x |
| 19 | Database data seeding | |
| 20 | Saga Pattern | |
| 21 | Unit of Work Pattern | |
| 22 | OpenAPI Documentation | x |
| 23 | Problem Details+json validation | x |
| 24 | Unit Test | x |
| 25 | Integration Test | x |
| 26 | CQRS | x |

### Dev tool dependencies

[sqlc](https://docs.sqlc.dev/en/latest/index.html)

```shell
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

[swag](https://github.com/swaggo/swag)

```shell
go install github.com/swaggo/swag/cmd/swag@latest
```
