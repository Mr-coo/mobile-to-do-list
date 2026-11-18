import { ActivityIndicator, FlatList, StyleSheet, Text, View } from 'react-native';
import { Task } from '../types';
import { TaskItem } from './TaskItem';

interface Props {
  tasks: Task[];
  isLoading: boolean;
  onPressTask?: (task: Task) => void;
}

export function TaskList({ tasks, isLoading, onPressTask }: Props) {
  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator testID="task-list-loading" />
      </View>
    );
  }

  return (
    <FlatList
      data={tasks}
      keyExtractor={(task) => String(task.id)}
      renderItem={({ item }) => <TaskItem task={item} onPress={onPressTask} />}
      contentContainerStyle={styles.content}
      ListEmptyComponent={<Text style={styles.empty}>No tasks found</Text>}
    />
  );
}

const styles = StyleSheet.create({
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 24 },
  content: { padding: 16 },
  empty: { textAlign: 'center', color: '#777', marginTop: 32 },
});
