-- Lists can be pinned to the menu bar menu (Settings); the Inbox starts pinned.
ALTER TABLE lists ADD COLUMN in_menu_bar INTEGER NOT NULL DEFAULT 0;
UPDATE lists SET in_menu_bar = 1 WHERE is_inbox = 1;
