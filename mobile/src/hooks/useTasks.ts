import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createTask, deleteTask, listTasks, updateTask } from '../api/tasks';
import { ListParams, TaskInput } from '../types';

export function useTaskList(params: ListParams) {
  return useQuery({
    queryKey: ['tasks', params],
    queryFn: () => listTasks(params),
    placeholderData: keepPreviousData,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: TaskInput) => createTask(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: TaskInput }) => updateTask(id, input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteTask(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}
