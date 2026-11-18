import { Pressable, StyleSheet, Text, View } from 'react-native';
import { TaskStatus } from '../types';
import { STATUS_LABELS } from './TaskItem';

interface Props {
  value?: TaskStatus;
  onChange: (status?: TaskStatus) => void;
}

const OPTIONS: { label: string; status?: TaskStatus }[] = [
  { label: 'All' },
  { label: STATUS_LABELS.todo, status: 'todo' },
  { label: STATUS_LABELS.in_progress, status: 'in_progress' },
  { label: STATUS_LABELS.done, status: 'done' },
];

export function StatusFilter({ value, onChange }: Props) {
  return (
    <View style={styles.row}>
      {OPTIONS.map((option) => {
        const selected = option.status === value;
        return (
          <Pressable
            key={option.label}
            accessibilityRole="button"
            accessibilityState={{ selected }}
            style={[styles.chip, selected && styles.chipSelected]}
            onPress={() => onChange(option.status)}
          >
            <Text style={selected ? styles.textSelected : styles.text}>{option.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  chip: {
    borderWidth: 1,
    borderColor: '#d4d4d4',
    borderRadius: 16,
    paddingVertical: 6,
    paddingHorizontal: 12,
    backgroundColor: '#fff',
  },
  chipSelected: { backgroundColor: '#2563eb', borderColor: '#2563eb' },
  text: { color: '#333' },
  textSelected: { color: '#fff', fontWeight: '600' },
});
