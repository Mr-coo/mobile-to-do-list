import { useState } from 'react';
import { Pressable, SafeAreaView, StyleSheet, Text, View } from 'react-native';
import { TaskFormModal } from '../components/TaskFormModal';
import { TaskList } from '../components/TaskList';
import { useCreateTask, useTaskList } from '../hooks/useTasks';
import { TaskInput } from '../types';

export function TasksScreen() {
  const [formVisible, setFormVisible] = useState(false);
  const { data, isLoading, error } = useTaskList();
  const createTask = useCreateTask();

  const closeForm = () => {
    setFormVisible(false);
    createTask.reset();
  };

  const handleCreate = (input: TaskInput) => {
    createTask.mutate(input, { onSuccess: closeForm });
  };

  return (
    <SafeAreaView style={styles.screen}>
      <View style={styles.header}>
        <Text style={styles.title}>Tasks</Text>
        <Pressable style={styles.addButton} onPress={() => setFormVisible(true)}>
          <Text style={styles.addText}>+ New task</Text>
        </Pressable>
      </View>
      {error ? <Text style={styles.error}>{error.message}</Text> : null}
      <TaskList tasks={data?.data ?? []} isLoading={isLoading} />
      <TaskFormModal
        visible={formVisible}
        onSubmit={handleCreate}
        onClose={closeForm}
        isSaving={createTask.isPending}
        error={createTask.error?.message}
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
  title: { fontSize: 24, fontWeight: '700' },
  addButton: { backgroundColor: '#2563eb', borderRadius: 6, paddingVertical: 8, paddingHorizontal: 14 },
  addText: { color: '#fff', fontWeight: '600' },
  error: { color: '#b91c1c', paddingHorizontal: 16 },
});
