CREATE TABLE lists (
  id         INTEGER PRIMARY KEY,
  name       TEXT    NOT NULL CHECK (length(trim(name)) > 0),
  is_inbox   INTEGER NOT NULL DEFAULT 0,
  position   REAL    NOT NULL,
  created_at TEXT    NOT NULL,
  updated_at TEXT    NOT NULL
);
CREATE UNIQUE INDEX one_inbox ON lists(is_inbox) WHERE is_inbox = 1;

CREATE TABLE tasks (
  id           INTEGER PRIMARY KEY,
  list_id      INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
  title        TEXT    NOT NULL CHECK (length(trim(title)) > 0),
  notes        TEXT    NOT NULL DEFAULT '',
  due_date     TEXT,                       -- 'YYYY-MM-DD' or NULL
  priority     INTEGER NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 3),
  completed_at TEXT,                       -- NULL = open
  position     REAL    NOT NULL,
  created_at   TEXT    NOT NULL,
  updated_at   TEXT    NOT NULL
);
CREATE INDEX tasks_by_list ON tasks(list_id, completed_at, position);
CREATE INDEX tasks_by_due  ON tasks(due_date) WHERE completed_at IS NULL;

INSERT INTO lists (name, is_inbox, position, created_at, updated_at)
VALUES ('Inbox', 1, 0,
        strftime('%Y-%m-%dT%H:%M:%SZ', 'now'),
        strftime('%Y-%m-%dT%H:%M:%SZ', 'now'));
