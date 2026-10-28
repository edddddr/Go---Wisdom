CREATE TABLE tasks (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT        NOT NULL CHECK (char_length(title) BETWEEN 1 AND 100),
    description TEXT        NOT NULL DEFAULT '' CHECK (char_length(description) <= 500),
    done        BOOLEAN     NOT NULL DEFAULT FALSE,
    priority    SMALLINT    NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_done ON tasks (done);

DROP TABLE IF EXISTS tasks;