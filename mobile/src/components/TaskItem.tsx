import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Task, TaskStatus } from '../types';

export const STATUS_LABELS: Record<TaskStatus, string> = {
  todo: 'To do',
  in_progress: 'In progress',
  done: 'Done',
};

interface Props {
  task: Task;
  onPress?: (task: Task) => void;
}

export function TaskItem({ task, onPress }: Props) {
  return (
    <Pressable style={styles.card} onPress={() => onPress?.(task)}>
      <View style={styles.header}>
        <Text style={styles.title}>{task.title}</Text>
        <Text style={styles.status}>{STATUS_LABELS[task.status]}</Text>
      </View>
      {task.description ? <Text style={styles.description}>{task.description}</Text> : null}
      {task.assignee ? <Text style={styles.assignee}>Assigned to {task.assignee}</Text> : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#fff',
    borderRadius: 8,
    padding: 12,
    marginBottom: 10,
    borderWidth: 1,
    borderColor: '#e2e2e2',
  },
  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  title: { fontSize: 16, fontWeight: '600', flex: 1, marginRight: 8 },
  status: { fontSize: 12, color: '#555' },
  description: { marginTop: 6, color: '#444' },
  assignee: { marginTop: 6, fontSize: 12, color: '#777' },
});
