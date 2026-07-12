SET search_path TO scheduler, public;

-- =====================================================
-- tasks table
-- =====================================================
DROP TRIGGER IF EXISTS trg_tasks_updated_at ON tasks;
DROP INDEX IF EXISTS idx_tasks_scheduled_at;
DROP TABLE IF EXISTS tasks;

-- =====================================================
-- Helper Functions
-- =====================================================
DROP FUNCTION IF EXISTS update_updated_at_column();

