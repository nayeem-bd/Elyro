# Elyro — Offline Todo App for macOS

A single-user, fully offline todo app for macOS (Apple Silicon, macOS 13+).
No accounts, no network, no sync. All data lives in one local SQLite file.

Stack and versions: see the approved version report (Wails v3.1.0, Go 1.27.4, Vue 3.5.45, TS 6.0.3,
Vite 8.3.8, Pinia 4.1.0, Vue Router 5.5.1, Tailwind 4.3.5, @lucide/vue 1.72.0, modernc.org/sqlite v1.62.0).

## v1 scope

**In:**
- Lists: a built-in **Inbox** plus user lists (create, rename, delete, reorder).
- Tasks: title, notes, due date, priority (none/low/medium/high), complete/uncomplete, delete, reorder, move between lists.
- Smart views: **Today** (due today or overdue), **Upcoming** (future due dates), **Completed**.
- Search across task titles and notes.
- Keyboard-first: quick add, navigate, complete, delete; native macOS menu with shortcuts.
- Follows system light/dark mode; remembers window size and position.
- Data safety: SQLite in WAL mode, automatic daily backup, JSON export/import.

**Out of v1** (could come later): subtasks, tags, recurring tasks, attachments, sync/iCloud, due-time reminders/notifications, multiple windows.

## Architecture

```
Vue (views, components)
  └─ Pinia stores (lists, tasks, ui)
       └─ generated Wails bindings (typed TS)
            └─ Go services (ListService, TaskService, DataService)
                 └─ internal/store (database/sql + modernc SQLite, embedded migrations)
                      └─ ~/Library/Application Support/Elyro/elyro.db
```

- **Go owns all data rules** (validation, ordering, dates). The frontend does display and input only.
- **No ORM, no query generator**: plain `database/sql`. The schema is small.
- **Migrations**: numbered `.sql` files embedded with `embed`, tracked with `PRAGMA user_version`, applied in one transaction at startup.
- **Ordering**: a `position REAL` column (fractional indexing), so a reorder updates one row.
- **Dates**: timestamps stored as UTC RFC 3339 text. Due dates stored as local `YYYY-MM-DD` (a due date is a calendar day, not an instant).
- **Search**: `LIKE` over title and notes (enough for a personal list; FTS5 only if it gets slow).
- **Offline guarantee**: frontend embedded in the binary, system font stack (no web fonts), no runtime network calls.

### Schema (migration 001)

```sql
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
```

Connection settings: `foreign_keys=ON`, `journal_mode=WAL`, `busy_timeout=5000`.

### Go services (bound to the frontend)

- `ListService`: `All`, `Create(name)`, `Rename(id, name)`, `Delete(id)` (Inbox cannot be deleted), `Move(id, beforeID, afterID)`
- `TaskService`: `ByList(listID, includeDone)`, `Today()`, `Upcoming()`, `Completed()`, `Search(q)`, `Create(input)`, `Update(id, input)`, `SetDone(id, done)`, `Delete(id)`, `Move(id, listID, beforeID, afterID)`
- `DataService`: `Export() (path)`, `Import(path)`, `Backup()`

### Frontend layout

- `router`: `/list/:id`, `/today`, `/upcoming`, `/completed`, `/search?q=`
- `stores`: `lists`, `tasks`, `ui` (selection, editing state)
- `components`: `Sidebar`, `TaskList`, `TaskRow`, `QuickAdd`, `TaskDetail` (notes, due date, priority), `EmptyState`

## Phases

Each phase ends with `wails3 task test` and `wails3 build` green, and a manual run of the app.

### Phase 1 — Foundation
- Remove the template demo: `GreetService`, `HelloWorld.vue`, the `time` event, background images, Inter font, `public/style.css`.
- App identity: name "Elyro", bundle identifier `com.elyro.app`, company and copyright in `build/config.yml`, then regenerate the plists.
- `internal/store`: open the DB in the Application Support folder, connection pragmas, embedded migrations, migration 001.
- Go tests: migrations apply on a fresh DB and are idempotent; schema constraints reject bad data.
- Empty app shell: window, Tailwind base, router with placeholder views, sidebar skeleton.

### Phase 2 — Core tasks
- `ListService` and `TaskService` CRUD (create, rename, complete, delete) with Go tests.
- Pinia stores calling the generated bindings.
- UI: sidebar of lists, task list, quick add (Enter to add), click to complete, inline title edit, delete.

### Phase 3 — Organization
- Due dates, priority, notes in a detail panel.
- Today, Upcoming, Completed views; search.
- Drag to reorder tasks and lists; move a task to another list.

### Phase 4 — macOS polish
- Native app menu and shortcuts (⌘N new task, ⌘⇧N new list, ⌘F search, ⌘⌫ delete, Space to complete, arrow navigation).
- System light/dark mode, window state persistence, empty states, focus rings and VoiceOver labels.

### Phase 5 — Data safety and release
- Automatic daily backup (SQLite `VACUUM INTO`, keep the last 7).
- JSON export/import with validation.
- `wails3 task package` and `.dmg`; app icon.

## Testing

- **Go**: table-driven tests for store and services against a temp-file DB (the real pragmas, not `:memory:`).
- **Frontend**: `vue-tsc` type checks in every build. No frontend test runner in v1; logic stays in Go.
- **Manual check per phase**: run the app, exercise the phase's features, quit and relaunch to confirm data persists.
