import { request } from './client';
import { ListParams, ListResponse, Task, TaskInput } from '../types';

export function listTasks(params: ListParams): Promise<ListResponse> {
  const query = new URLSearchParams({ page: String(params.page), limit: String(params.limit) });
  if (params.keyword) query.set('keyword', params.keyword);
  if (params.status) query.set('status', params.status);
  return request(`/tasks?${query.toString()}`);
}

export function createTask(input: TaskInput): Promise<{ data: Task }> {
  return request('/tasks', { method: 'POST', body: JSON.stringify(input) });
}

export function updateTask(id: number, input: TaskInput): Promise<{ data: Task }> {
  return request(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(input) });
}

export function deleteTask(id: number): Promise<void> {
  return request(`/tasks/${id}`, { method: 'DELETE' });
}
