-- name: CreateSagaInstance :one
INSERT INTO saga_instances (id, status, payload, created_at, updated_at)
VALUES (@id, @status, @payload, now(), now())
RETURNING *;

-- name: GetSagaInstance :one
SELECT * FROM saga_instances WHERE id = $1;

-- name: TransitionSagaInstanceStatus :one
UPDATE saga_instances
SET    status = @new_status, updated_at = now()
WHERE  id = @id AND status = @current_status
RETURNING id;

-- name: CreateSagaStep :exec
INSERT INTO saga_steps (saga_id, step_name, status, created_at, updated_at)
VALUES (@saga_id, @step_name, @status, now(), now());

-- name: TransitionSagaStepStatus :one
UPDATE saga_steps
SET    status = @new_status, updated_at = now()
WHERE  saga_id = @saga_id AND step_name = @step_name AND status = @current_status
RETURNING id;
