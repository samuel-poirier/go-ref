-- name: FindFirstOutboxMessageByScheduledTime :one
SELECT * from outbox
ORDER BY scheduled_at
OFFSET 0
LIMIT 1;

-- name: CreateOutboxMessage :exec
INSERT INTO outbox (
    id,
    created_at,
    scheduled_at,
    metadata,
    payload,
    times_attempted
) VALUES ($1, now(), $2, $3, $4, 0);

-- name: IncrementOutboxMessageTimesAttemptedById :exec
UPDATE outbox set times_attempted = times_attempted + 1
where id = $1;

-- name: DeleteOutboxMessageById :exec
DELETE from outbox where id = $1;
