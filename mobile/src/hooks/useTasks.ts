import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createTask, listTasks } from '../api/tasks';
import { TaskInput } from '../types';

export function useTaskList() {
  return useQuery({ queryKey: ['tasks'], queryFn: listTasks });
}

export function useCreateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: TaskInput) => createTask(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}
