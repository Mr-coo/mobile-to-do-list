ALTER TABLE tasks
  DROP INDEX uq_tasks_active_title,
  DROP COLUMN active_title,
  ADD UNIQUE INDEX uq_tasks_title (title);
