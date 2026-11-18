import { useCallback, useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Pagination } from '../components/Pagination';
import { SearchBar } from '../components/SearchBar';
import { StatusFilter } from '../components/StatusFilter';
import { TaskFormModal } from '../components/TaskFormModal';
import { TaskList } from '../components/TaskList';
import { useCreateTask, useDeleteTask, useTaskList, useUpdateTask } from '../hooks/useTasks';
import { Task, TaskInput, TaskStatus } from '../types';

const PAGE_SIZE = 10;

export function TasksScreen() {
  const [formVisible, setFormVisible] = useState(false);
  const [editing, setEditing] = useState<Task | undefined>();
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<TaskStatus | undefined>();
  const [page, setPage] = useState(1);
  const { data, isLoading, isFetching, isError, refetch } = useTaskList({
    keyword,
    status,
    page,
    limit: PAGE_SIZE,
  });
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();
  const deleteTask = useDeleteTask();

  const handleSearch = useCallback((value: string) => {
    setKeyword(value);
    setPage(1);
  }, []);

  const handleStatusChange = (value?: TaskStatus) => {
    setStatus(value);
    setPage(1);
  };

  const closeForm = () => {
    setFormVisible(false);
    setEditing(undefined);
    createTask.reset();
    updateTask.reset();
    deleteTask.reset();
  };

  const openEdit = (task: Task) => {
    setEditing(task);
    setFormVisible(true);
  };

  const handleSubmit = (input: TaskInput) => {
    if (editing) {
      updateTask.mutate({ id: editing.id, input }, { onSuccess: closeForm });
    } else {
      createTask.mutate(input, { onSuccess: closeForm });
    }
  };

  const handleDelete = () => {
    if (editing) deleteTask.mutate(editing.id, { onSuccess: closeForm });
  };

  const saveMutation = editing ? updateTask : createTask;

  return (
    <SafeAreaView style={styles.screen}>
      <View style={styles.header}>
        <View style={styles.titleRow}>
          <Text style={styles.title}>Tasks</Text>
          {isFetching && !isLoading ? <ActivityIndicator testID="refetching" size="small" /> : null}
        </View>
        <Pressable style={styles.addButton} onPress={() => setFormVisible(true)}>
          <Text style={styles.addText}>+ New task</Text>
        </Pressable>
      </View>
      <View style={styles.controls}>
        <SearchBar onSearch={handleSearch} />
        <StatusFilter value={status} onChange={handleStatusChange} />
      </View>
      {isError ? (
        <View style={styles.errorBox}>
          <Text style={styles.error}>Couldn't load tasks</Text>
          <Pressable style={styles.retryButton} onPress={() => refetch()}>
            <Text style={styles.addText}>Retry</Text>
          </Pressable>
        </View>
      ) : (
        <TaskList tasks={data?.data ?? []} isLoading={isLoading} onPressTask={openEdit} />
      )}
      <Pagination
        page={page}
        totalPages={data?.meta.total_pages ?? 1}
        onChange={setPage}
        disabled={isFetching}
      />
      <TaskFormModal
        visible={formVisible}
        task={editing}
        onSubmit={handleSubmit}
        onDelete={handleDelete}
        onClose={closeForm}
        isSaving={saveMutation.isPending || deleteTask.isPending}
        error={deleteTask.error ?? saveMutation.error}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: '#f6f6f6' },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    paddingTop: 32,
  },
  titleRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  controls: { paddingHorizontal: 16, gap: 8 },
  title: { fontSize: 24, fontWeight: '700' },
  addButton: { backgroundColor: '#2563eb', borderRadius: 6, paddingVertical: 8, paddingHorizontal: 14 },
  addText: { color: '#fff', fontWeight: '600' },
  errorBox: { flex: 1, justifyContent: 'center', alignItems: 'center', gap: 12 },
  error: { color: '#b91c1c' },
  retryButton: { backgroundColor: '#2563eb', borderRadius: 6, paddingVertical: 8, paddingHorizontal: 14 },
});
