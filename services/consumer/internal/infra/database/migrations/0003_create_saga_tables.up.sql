CREATE TABLE IF NOT EXISTS saga_instances (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    status     VARCHAR(50)  NOT NULL,
    payload    JSONB        NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS saga_steps (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id    UUID         NOT NULL REFERENCES saga_instances(id),
    step_name  VARCHAR(100) NOT NULL,
    status     VARCHAR(50)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (saga_id, step_name)
);

CREATE INDEX idx_saga_instances_status ON saga_instances (status);
CREATE INDEX idx_saga_steps_saga_id    ON saga_steps (saga_id);
