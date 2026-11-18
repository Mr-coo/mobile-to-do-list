export type TaskStatus = 'todo' | 'in_progress' | 'done';

export interface Task {
  id: number;
  title: string;
  description: string | null;
  status: TaskStatus;
  assignee: string | null;
  created_at: string;
  updated_at: string;
}

export interface TaskInput {
  title: string;
  description: string | null;
  status: TaskStatus;
  assignee: string | null;
}
