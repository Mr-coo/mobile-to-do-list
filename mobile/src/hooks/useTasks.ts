import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createTask, listTasks } from '../api/tasks';
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
