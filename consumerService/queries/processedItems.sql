-- name: CountAllProcessedItems :one
SELECT COUNT(*) FROM processed_items;

-- name: FindLastProcessedItem :one
SELECT * from processed_items order by processed_at desc limit 1;

-- name: CreateProcessedItem :exec
INSERT INTO processed_items (processed_data, processed_at) VALUES (sqlc.arg(processed_data)::VARCHAR(255), now());

-- name: FindProcessedItemsWithPaging :many
SELECT * from processed_items
order by processed_at
OFFSET $1
LIMIT $2;

