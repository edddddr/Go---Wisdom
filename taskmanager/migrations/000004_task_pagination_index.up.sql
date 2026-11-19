CREATE INDEX idx_tasks_owner_id_id ON tasks (owner_id, id DESC);


DROP INDEX idx_tasks_owner_id;