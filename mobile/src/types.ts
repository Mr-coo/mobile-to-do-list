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

export interface ListParams {
  keyword?: string;
  status?: TaskStatus;
  page: number;
  limit: number;
}

export interface ListMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ListResponse {
  data: Task[];
  meta: ListMeta;
}
