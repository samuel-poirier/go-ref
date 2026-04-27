# RFC: Orchestrated Saga Pattern — Item Processing

Author: @samuel-poirier   Status: Draft   Date: 2026-04-26

## Summary

Add a concrete, orchestrated saga example to the consumer service demonstrating how
long-running multi-step workflows are coordinated with strong resilience guarantees.
An HTTP POST triggers a saga that drives an item through `pending → processing →
processed-succeeded | processed-failed`. Saga state is persisted in two new Postgres
tables (`saga_instances`, `saga_steps`). All cross-service messaging goes through the
existing outbox pattern, so every state transition and command dispatch is atomic with
its database write. A compensating transaction is executed on failure to demonstrate
rollback choreography.

## Motivation

1. The existing codebase demonstrates at-least-once delivery via the outbox pattern, but
   has no example of multi-step workflow coordination across multiple consumers.
2. A central-coordinator saga is the most common pattern teams reach for when building
   order fulfillment, payment processing, or provisioning flows in Go — this reference
   makes it copy-pasteable.
3. Using the existing outbox infrastructure avoids adding Temporal, Conductor, or any
   external orchestration dependency, keeping the dependency footprint minimal.

## Why not alternatives

- **Choreography (event-driven, no central coordinator):** Each participant would
  react to the previous participant's event with no central place to see the saga's
  current state. Harder to reason about, harder to add compensating logic, and the
  flow lives scattered across consumers. Rejected in favour of an explicit orchestrator
  that owns the state machine.
- **In-memory saga state only:** The orchestrator could hold current step state in RAM
  and only persist the final result. Rejected because a process restart would lose all
  in-flight sagas, undermining the resilience goal. Every state transition must survive
  a crash.
- **External saga library (go-saga, Temporal, Conductor):** Each adds a non-trivial
  external dependency and an unfamiliar operational surface. The goal is to show the
  pattern built from first principles on top of what already exists. Rejected.

## Design

### Saga lifecycle

```
[HTTP POST /api/v1/saga/items]
        │
        ▼  (UoW: INSERT saga_instances + outbox)
   pending  ──────────────────────────────────────────►  StartItemSagaCommand
        │                                                        │ (orchestrator)
        ▼  (UoW: INSERT saga_steps(process_item,pending)         │
   processing ◄──────────────────────────────────────────────────┘
        │        + outbox: ProcessItemCommand)
        │
        ├── success ──► (UoW: step→succeeded, instance→processed-succeeded)
        │                log "saga {id} completed successfully"
        │
        └── failure ──► (UoW: step→failed, INSERT saga_steps(compensate,pending)
                              + outbox: CompensateProcessItemCommand)
                                │
                                ▼  (activity: log + outbox: ProcessItemCompensatedEvent)
                         (UoW: compensate_step→compensated, instance→processed-failed)
                          log "saga {id} completed with failure, compensation applied"
```

### Schema

Migration file: `services/consumer/internal/infra/database/migrations/0003_create_saga_tables.up.sql`

```sql
CREATE TABLE IF NOT EXISTS saga_instances (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status     VARCHAR(50)  NOT NULL,
    payload    JSONB        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS saga_steps (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id    UUID         NOT NULL REFERENCES saga_instances(id),
    step_name  VARCHAR(100) NOT NULL,
    status     VARCHAR(50)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (saga_id, step_name)
);

CREATE INDEX idx_saga_instances_status ON saga_instances (status);
CREATE INDEX idx_saga_steps_saga_id    ON saga_steps (saga_id);
```

Down migration: `0003_create_saga_tables.down.sql`

```sql
DROP TABLE IF EXISTS saga_steps;
DROP TABLE IF EXISTS saga_instances;
```

#### Status values

| Table            | Status values                                                           |
|------------------|-------------------------------------------------------------------------|
| saga_instances   | `pending`, `processing`, `processed-succeeded`, `processed-failed`      |
| saga_steps       | `pending`, `processing`, `succeeded`, `failed`, `compensating`, `compensated` |

#### Step name constants (Go)

```go
// In pkg/events/saga.go or a new constants file — visible to commands package
const (
    StepProcessItem          = "process_item"
    StepCompensateProcessItem = "compensate_process_item"
)
```

### sqlc queries

File: `services/consumer/queries/saga.sql`

```sql
-- name: CreateSagaInstance :one
INSERT INTO saga_instances (id, status, payload, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
RETURNING *;

-- name: GetSagaInstance :one
SELECT * FROM saga_instances WHERE id = $1;

-- name: TransitionSagaInstanceStatus :one
UPDATE saga_instances
SET    status = $3, updated_at = now()
WHERE  id = $1 AND status = $2
RETURNING id;

-- name: CreateSagaStep :exec
INSERT INTO saga_steps (saga_id, step_name, status, created_at, updated_at)
VALUES ($1, $2, $3, now(), now());

-- name: TransitionSagaStepStatus :one
UPDATE saga_steps
SET    status = $4, updated_at = now()
WHERE  saga_id = $1 AND step_name = $2 AND status = $3
RETURNING id;
```

Run `sqlc generate` after adding this file. The generated code lands in
`services/consumer/internal/repository/saga.sql.go`.

`TransitionSagaInstanceStatus` and `TransitionSagaStepStatus` return `pgx.ErrNoRows`
when the WHERE clause matches no row. This is the deduplication signal: the message
has already been processed (the status is no longer what was expected).

### Events and commands

File: `pkg/events/saga.go`

```go
package events

// HTTP → Outbox → Orchestrator
type StartItemSagaCommand struct {
    SagaID string `json:"saga_id"`
    Data   string `json:"data"`
}

// Orchestrator → Outbox → ProcessItem activity
type ProcessItemCommand struct {
    SagaID string `json:"saga_id"`
    Data   string `json:"data"`
}

// Orchestrator → Outbox → CompensateProcessItem activity
type CompensateProcessItemCommand struct {
    SagaID string `json:"saga_id"`
}

// ProcessItem activity → Outbox → Orchestrator
type ItemProcessedSucceededEvent struct {
    SagaID        string `json:"saga_id"`
    ProcessedData string `json:"processed_data"`
}

// ProcessItem activity → Outbox → Orchestrator
type ItemProcessedFailedEvent struct {
    SagaID  string `json:"saga_id"`
    Reason  string `json:"reason"`
}

// CompensateProcessItem activity → Outbox → Orchestrator
type ProcessItemCompensatedEvent struct {
    SagaID string `json:"saga_id"`
}
```

Queue names are derived from the struct type name via reflection (same as the existing
`repository.NewCreateOutboxMessageParams` pattern). No manual queue name registration
is required.

### Deduplication

The `UNIQUE (saga_id, step_name)` constraint on `saga_steps` is the dedup anchor.
The pattern used throughout:

1. **Before** the UoW transaction: do any side-effectful work (e.g., stubbed sleep).
   Keep transactions short.
2. **Inside** the UoW: call `TransitionSaga*Status`. If `pgx.ErrNoRows` is returned,
   return the sentinel `commands.ErrDuplicateMessage`.
3. **In the consumer handler**: check `errors.Is(err, commands.ErrDuplicateMessage)` →
   `msg.Ack()` and return. Any other error → `msg.Nack(true)`.

```go
// commands package
var ErrDuplicateMessage = errors.New("saga: duplicate message, already processed")
```

### Commands interface additions

File: `services/consumer/internal/app/service/commands/commands.go` — extend `Commands`:

```go
type Commands interface {
    CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error
    // Saga orchestrator handlers
    SagaHandleStart(ctx context.Context, cmd SagaHandleStartCommand) error
    SagaHandleSucceeded(ctx context.Context, cmd SagaHandleSucceededCommand) error
    SagaHandleFailed(ctx context.Context, cmd SagaHandleFailedCommand) error
    SagaHandleCompensated(ctx context.Context, cmd SagaHandleCompensatedCommand) error
    // Saga activity handlers
    SagaProcessItem(ctx context.Context, cmd SagaProcessItemCommand) error
    SagaCompensateProcessItem(ctx context.Context, cmd SagaCompensateProcessItemCommand) error
    // HTTP trigger
    SagaCreate(ctx context.Context, cmd SagaCreateCommand) (uuid.UUID, error)
}
```

### HTTP trigger

**New handler package:** `services/consumer/internal/app/api/saga/`

Files:
- `handler.go` — struct + constructor (mirror `api/processed/handler.go`)
- `start_item_saga.go` — POST handler

Endpoint: `POST /api/v1/saga/items`

Request body:
```json
{ "data": "some item data" }
```

Response `202 Accepted`:
```json
{ "saga_id": "550e8400-e29b-41d4-a716-446655440000" }
```

Handler logic (`SagaCreate` command):
1. Generate `sagaID = uuid.New()`.
2. UoW:
   - `CreateSagaInstance(sagaID, "pending", {"data": cmd.Data})`
   - Write `StartItemSagaCommand{SagaID: sagaID, Data: cmd.Data}` to outbox.
3. Return `sagaID` → HTTP 202.

The HTTP handler does not wait for the saga to complete. The caller polls or ignores.

### Orchestrator consumers

Four consumers, all in `services/consumer/internal/app/consumers/saga/orchestrator/`.
Each follows the same shape as `consumers/processed/consumer.go`.

#### 1. `start_consumer.go` — queue `StartItemSagaCommand`

`SagaHandleStart` command logic (inside UoW):
1. `TransitionSagaInstanceStatus(sagaID, "pending", "processing")` — `ErrNoRows` → return `ErrDuplicateMessage`.
2. `CreateSagaStep(sagaID, StepProcessItem, "pending")`.
3. Write `ProcessItemCommand{SagaID, Data}` to outbox.
4. Log `slog.Info("saga started processing", "saga_id", sagaID)`.

#### 2. `succeeded_consumer.go` — queue `ItemProcessedSucceededEvent`

`SagaHandleSucceeded` command logic (inside UoW):
1. `TransitionSagaStepStatus(sagaID, StepProcessItem, "processing", "succeeded")` — `ErrNoRows` → `ErrDuplicateMessage`.
2. `TransitionSagaInstanceStatus(sagaID, "processing", "processed-succeeded")`.
3. Log `slog.Info("saga completed successfully", "saga_id", sagaID)`.

#### 3. `failed_consumer.go` — queue `ItemProcessedFailedEvent`

`SagaHandleFailed` command logic (inside UoW):
1. `TransitionSagaStepStatus(sagaID, StepProcessItem, "processing", "failed")` — `ErrNoRows` → `ErrDuplicateMessage`.
2. `CreateSagaStep(sagaID, StepCompensateProcessItem, "pending")`.
3. Write `CompensateProcessItemCommand{SagaID}` to outbox.
4. Log `slog.Warn("saga step failed, dispatching compensation", "saga_id", sagaID, "reason", reason)`.

Note: `saga_instances.status` stays `"processing"` until compensation completes.

#### 4. `compensated_consumer.go` — queue `ProcessItemCompensatedEvent`

`SagaHandleCompensated` command logic (inside UoW):
1. `TransitionSagaStepStatus(sagaID, StepCompensateProcessItem, "compensating", "compensated")` — `ErrNoRows` → `ErrDuplicateMessage`.
2. `TransitionSagaInstanceStatus(sagaID, "processing", "processed-failed")`.
3. Log `slog.Info("saga completed with failure, compensation applied", "saga_id", sagaID)`.

### Activity consumers

Two consumers in `services/consumer/internal/app/consumers/saga/activities/`.

#### `process_item_consumer.go` — queue `ProcessItemCommand`

`SagaProcessItem` command logic:

```
// BEFORE UoW — keep transaction short
outcome := simulateWork()   // time.Sleep(rand 1-3s) + rand success/failure
processedData := "processed:" + cmd.Data  // if success
reason := "simulated failure"             // if failure

// INSIDE UoW
TransitionSagaStepStatus(sagaID, StepProcessItem, "pending", "processing")
  → ErrNoRows: return ErrDuplicateMessage
if outcome == success:
    Write ItemProcessedSucceededEvent{SagaID, ProcessedData} to outbox
    Log Info "process_item activity succeeded"
else:
    Write ItemProcessedFailedEvent{SagaID, Reason} to outbox
    Log Warn "process_item activity failed"
```

`simulateWork()` implementation:
```go
func simulateWork() bool {
    time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
    return rand.Float32() > 0.4  // 60% success rate
}
```

#### `compensate_item_consumer.go` — queue `CompensateProcessItemCommand`

`SagaCompensateProcessItem` command logic:

```
// BEFORE UoW
time.Sleep(500 * time.Millisecond)  // simulate compensation work

// INSIDE UoW
TransitionSagaStepStatus(sagaID, StepCompensateProcessItem, "pending", "compensating")
  → ErrNoRows: return ErrDuplicateMessage
Write ProcessItemCompensatedEvent{SagaID} to outbox
Log Info "compensate_process_item activity executed"
```

### Wiring

**`app.go`** — append six new handlers after existing two:

```go
consumerHandlers = append(consumerHandlers, sagaorchestrator.NewStartConsumer(service, ctx))
consumerHandlers = append(consumerHandlers, sagaorchestrator.NewSucceededConsumer(service, ctx))
consumerHandlers = append(consumerHandlers, sagaorchestrator.NewFailedConsumer(service, ctx))
consumerHandlers = append(consumerHandlers, sagaorchestrator.NewCompensatedConsumer(service, ctx))
consumerHandlers = append(consumerHandlers, sagaactivities.NewProcessItemConsumer(service, ctx))
consumerHandlers = append(consumerHandlers, sagaactivities.NewCompensateItemConsumer(service, ctx))
```

**`routes.go`** — add one new route in `v1`:

```go
sagaHandler := saga.NewHandler(service)
v1.HandleFunc("POST /api/v1/saga/items", sagaHandler.StartItemSaga)
```

### File map

| Action   | Path |
|----------|------|
| CREATE   | `services/consumer/internal/infra/database/migrations/0003_create_saga_tables.up.sql` |
| CREATE   | `services/consumer/internal/infra/database/migrations/0003_create_saga_tables.down.sql` |
| CREATE   | `services/consumer/queries/saga.sql` |
| GENERATE | `services/consumer/internal/repository/saga.sql.go` (via `sqlc generate`) |
| CREATE   | `pkg/events/saga.go` |
| MODIFY   | `services/consumer/internal/app/service/commands/commands.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_create.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_handle_start.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_handle_succeeded.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_handle_failed.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_handle_compensated.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_process_item.go` |
| CREATE   | `services/consumer/internal/app/service/commands/saga_compensate_process_item.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/orchestrator/start_consumer.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/orchestrator/succeeded_consumer.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/orchestrator/failed_consumer.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/orchestrator/compensated_consumer.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/activities/process_item_consumer.go` |
| CREATE   | `services/consumer/internal/app/consumers/saga/activities/compensate_item_consumer.go` |
| CREATE   | `services/consumer/internal/app/api/saga/handler.go` |
| CREATE   | `services/consumer/internal/app/api/saga/start_item_saga.go` |
| MODIFY   | `services/consumer/internal/app/app.go` |
| MODIFY   | `services/consumer/internal/app/routes.go` |

## Risks

1. **Long-running stub transaction:** The activity calls `time.Sleep` before opening the
   UoW. If an agent moves the sleep inside the transaction, the DB connection is held
   for 1-3 seconds per message. **Detection:** `pgxTracer` logs slow queries; visible
   in OpenTelemetry traces. **Mitigation:** the spec explicitly requires sleep to happen
   before the UoW block (see activity pseudocode).

2. **Stuck sagas (processing forever):** If the consumer crashes between the step
   transition commit and the result event being published from the outbox, the step is
   `processing` and the outbox entry is pending. The outbox reader delivers the result
   event on recovery, so this is self-healing. No timeout/reaper is needed for this
   reference.

3. **Duplicate `CreateSagaStep` on orchestrator retry:** If the orchestrator crashes
   after committing `TransitionSagaInstanceStatus` but before the ACK is sent, the
   broker redelivers `StartItemSagaCommand`. The transition query returns `ErrNoRows`
   (instance is now `processing`, not `pending`), triggering `ErrDuplicateMessage` →
   ACK. The already-inserted `saga_steps` row is untouched. Safe.

4. **`saga_steps` UNIQUE conflict on CREATE:** `CreateSagaStep` uses a plain `INSERT`,
   not `ON CONFLICT DO NOTHING`. If the orchestrator retries after a partial failure it
   could hit a unique violation before reaching the `TransitionSagaInstanceStatus`
   check. **Mitigation:** Always call `TransitionSagaInstanceStatus` first in the UoW;
   a successful transition guarantees this is the first time through. Order matters —
   the spec mandates it.

5. **No saga_instances GET endpoint:** The HTTP caller receives a `saga_id` but cannot
   query saga status without directly hitting the DB. This is acceptable for a reference
   implementation. A `GET /api/v1/saga/items/{id}` endpoint is the obvious next step if
   this pattern is promoted to production.

## Acceptance criteria for "done"

1. `POST /api/v1/saga/items` with `{"data": "test"}` returns HTTP 202 with a valid UUID
   `saga_id`.
2. Consumer service logs show the full happy-path transition sequence:
   `"saga started processing"` → `"process_item activity succeeded"` →
   `"saga completed successfully"`.
3. Consumer service logs show the failure + compensation sequence on a failed run:
   `"saga step failed, dispatching compensation"` →
   `"compensate_process_item activity executed"` →
   `"saga completed with failure, compensation applied"`.
4. After a completed saga (either terminal state), `saga_instances.status` is either
   `processed-succeeded` or `processed-failed` in the DB — never `pending` or
   `processing`.
5. All `saga_steps` rows for a completed saga are in a terminal status
   (`succeeded`, `failed`, or `compensated`). No step is left in `pending` or
   `processing`.
6. Sending the same `StartItemSagaCommand` twice (simulate broker redelivery) results
   in one saga, not two — verified by checking exactly one `saga_instances` row and the
   second delivery ACKing without error.
7. Parallel sagas: issue 3 concurrent `POST /api/v1/saga/items` calls; all three
   complete independently with no state bleed between instances.

## Open questions

- None. All design decisions were resolved during the spec interview.
