SET search_path TO scheduler, public;

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
    task text NOT NULL,
    scheduled_at timestamp NOT NULL,
    picked_at timestamp,
    started_at timestamp,
    completed_at timestamp,
    failed_at timestamp,
    error text,
    created_at timestamp DEFAULT NOW(),
    updated_at timestamp DEFAULT NOW()
);

CREATE TRIGGER trg_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_tasks_scheduled_at ON tasks(scheduled_at);

