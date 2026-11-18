import { request } from './client';
import { Task, TaskInput } from '../types';

export function listTasks(): Promise<{ data: Task[] }> {
  return request('/tasks');
}

export function createTask(input: TaskInput): Promise<{ data: Task }> {
  return request('/tasks', { method: 'POST', body: JSON.stringify(input) });
}
