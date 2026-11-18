-- Title must be unique among active tasks only, so a soft-deleted task frees its title.
-- MySQL has no partial unique index; a unique index ignores NULLs, so this generated
-- column is the title for active rows and NULL for deleted ones.
ALTER TABLE tasks
  DROP INDEX uq_tasks_title,
  ADD COLUMN active_title VARCHAR(255)
    GENERATED ALWAYS AS (IF(deleted_at IS NULL, title, NULL)) STORED,
  ADD UNIQUE INDEX uq_tasks_active_title (active_title);
