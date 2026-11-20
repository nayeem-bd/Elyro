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
- Runs in the background: menu bar icon with quick actions, optional launch at login.

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
- **Ordering**: a `position` column; a move shifts the rows after the drop point down by one (lists hold tens of rows, so simple beats clever).
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

- `ListService`: `All`, `Create(name)`, `Rename(id, name)`, `ShowInMenuBar(id, on)`, `Delete(id)` (Inbox cannot be deleted), `Move(id, beforeID, afterID)`
- `TaskService`: `ByList(listID, includeDone)`, `Counts()`, `Today()`, `Upcoming()`, `Completed()`, `Search(q)`, `Create(input)`, `Update(id, input)`, `SetDone(id, done)`, `Delete(id)`, `Move(id, listID, beforeID, afterID)`
- `DataService`: `Export(path)`, `Import(path)` (replaces everything, after a `Backups/before-import-<time>.db` snapshot, last 5 kept); the daily backup runs in Go at launch

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

### Phase 6 — Background and menu bar
Elyro keeps running when its window is closed, can start at login, and has a menu bar icon.
All of it uses Wails v3.1.0 built-ins; no new dependencies.

**Runs in the background**
- Closing the window hides it instead of quitting (`WindowClosing` hook → `Hide()` + cancel);
  `ApplicationShouldTerminateAfterLastWindowClosed` becomes `false`. Wails shows a hidden window
  when the Dock icon is clicked (unless the icon is hidden, see **Dock icon** below).
- ⌘Q, Quit in the menu bar menu, and logout still quit for real; window state is still saved on quit.
- Accepted cost: the hidden window keeps its web view in memory, so reopening is instant.

**Menu bar icon** (`app.SystemTray`)
- A monochrome ✓ template icon (`SetTemplateIcon`, 18×18 @1x/@2x PNG, embedded), so macOS tints it
  for light/dark menu bars.
- Menu items reuse the existing `send()` → frontend `commands` path, after `win.Show()` + `Focus()`:
  - **Quick Add Task…** (⌥⌘N): the existing `newTask` command.
  - **Today (n)**: sends the view path `/today` (the frontend routes any `/…` menu event). *n* comes
    from `TaskService.Counts()`, counted when the menu is rebuilt, so it is never stale and needs no
    timer or write hooks.
  - **Lists picked in Settings** (`/settings`, gear in the sidebar): "Inbox (n)", "Work (n)"… with open
    counts (`TaskService.Counts`, one grouped query; the sidebar shows the same counts); a `lists.in_menu_bar` flag (migration 002, Inbox on by default, kept in exports).
    The menu is rebuilt on every open (`Clear` → refill → `SetMenu` → `OpenMenu`).
  - **Launch at Login** (checkbox, only here), **Settings…**.
  - **Quit Elyro**.
- Don't use `AttachWindow`/`ShowWindow`/`ToggleWindow`: they move the window under the tray
  icon and would overwrite its saved position.

**System-wide shortcut**
- ⌥⌘N = Quick Add from any app (`app.GlobalShortcut.Register`, Carbon hot keys, no Accessibility
  permission), the same function as the menu item. If another app owns ⌥⌘N, log it and keep the
  menu item working.

**Launch at login** (`app.Autostart`, which uses `SMAppService` on macOS 13+)
- The checkbox shows the real registration status. If enabling fails (e.g. the user blocked Elyro in
  System Settings › General › Login Items), a native error dialog says so (the window may be hidden,
  so the in-window error banner could go unseen).
- **Start hidden**: when Launch at Login is on, the window is created hidden. `SMAppService` can't
  pass a "launched at login" flag, so a manual cold start is also hidden; a Dock or menu bar click
  shows the window.

**Midnight due labels** (separate small fix)
- Re-render the due labels when the window is shown or focused, so "Today/Tomorrow" are right
  after midnight without a timer running in the background.

**Checks**
- Go tests: `TestCounts` (Today and per-list open counts, pinned clock), `TestSettings`.
- Manual, using the **installed** app (login items need the real `/Applications/Elyro.app`):
  - close the window and reopen it from the Dock and from the menu bar;
  - ⌥⌘N from another app;
  - toggle Launch at Login, log out and in, and check Elyro is running hidden;
  - the Today count after adding a task;
  - the due labels after midnight.
- Verified in the Wails source: a tray click handler runs even with a menu set, so the count is
  refreshed in `OnClick` and the menu opened with `OpenMenu()`.

**Dock icon** (Settings › Dock): `settings.json` next to the database (`SettingsService`). Kept out
of the database on purpose: it is a per-Mac preference, so backups and exports don't carry it. Hidden = `ActivationPolicyAccessory` at launch, switched live with
the Wails dock service (`dock.New().HideAppIcon/ShowAppIcon`).

**Out of scope**: a separate floating quick-add panel and a user-configurable shortcut.

## Testing

- **Go**: table-driven tests for store and services against a temp-file DB (the real pragmas, not `:memory:`).
- **Frontend**: `vue-tsc` type checks in every build. No frontend test runner in v1; logic stays in Go.
- **Manual check per phase**: run the app, exercise the phase's features, quit and relaunch to confirm data persists.
