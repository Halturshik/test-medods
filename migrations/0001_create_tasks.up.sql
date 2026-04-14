CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    repeat_config JSONB,
    generated_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_tasks_parent_due 
ON tasks (parent_id, due_date) 
WHERE parent_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks (parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks (due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_status ON tasks (parent_id, status);