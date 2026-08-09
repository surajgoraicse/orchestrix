SET search_path TO scheduler, public;

CREATE TYPE task_status AS ENUM ('PENDING', 'DISPATCHED', 'COMPLETED', 'FAILED');

-- =====================================================
-- Helper Functions
-- =====================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- =====================================================
-- tasks table
-- =====================================================
CREATE TABLE IF NOT EXISTS tasks(
    id uuid PRIMARY KEY DEFAULT uuidv7(),

    -- task routing
    task_type varchar(255) NOT NULL,
    payload jsonb NOT NULL,

    -- task state
    status task_status NOT NULL DEFAULT 'PENDING',

    -- retry states
    max_retries int NOT NULL DEFAULT 3,
    attempt_count int NOT NULL DEFAULT 0,

    -- task timeline
    scheduled_at timestamptz NOT NULL,
    picked_at timestamptz,
    started_at timestamptz,
    completed_at timestamptz,
    failed_at timestamptz,

    -- error tracking
    error text,

    -- audit trails
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW()
);

CREATE TRIGGER trg_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_tasks_scheduled_at ON tasks(scheduled_at);

