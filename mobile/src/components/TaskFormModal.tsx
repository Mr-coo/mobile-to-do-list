import { useEffect, useState } from 'react';
import { Modal, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { Task, TaskInput, TaskStatus } from '../types';
import { STATUS_LABELS } from './TaskItem';

interface Props {
  visible: boolean;
  task?: Task;
  onSubmit: (input: TaskInput) => void;
  onClose: () => void;
  isSaving: boolean;
  error?: string;
}

const STATUSES: TaskStatus[] = ['todo', 'in_progress', 'done'];

export function TaskFormModal({ visible, task, onSubmit, onClose, isSaving, error }: Props) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState<TaskStatus>('todo');
  const [assignee, setAssignee] = useState('');

  useEffect(() => {
    if (visible) {
      setTitle(task?.title ?? '');
      setDescription(task?.description ?? '');
      setStatus(task?.status ?? 'todo');
      setAssignee(task?.assignee ?? '');
    }
  }, [visible, task]);

  const submit = () => {
    onSubmit({
      title: title.trim(),
      description: description.trim() || null,
      status,
      assignee: assignee.trim() || null,
    });
  };

  return (
    <Modal visible={visible} animationType="slide" onRequestClose={onClose}>
      <ScrollView contentContainerStyle={styles.container}>
        <Text style={styles.heading}>{task ? 'Edit task' : 'New task'}</Text>

        <Text style={styles.label}>Title</Text>
        <TextInput style={styles.input} value={title} onChangeText={setTitle} placeholder="Title" />

        <Text style={styles.label}>Description</Text>
        <TextInput
          style={[styles.input, styles.multiline]}
          value={description}
          onChangeText={setDescription}
          placeholder="Description"
          multiline
        />

        <Text style={styles.label}>Status</Text>
        <View style={styles.chips}>
          {STATUSES.map((value) => (
            <Pressable
              key={value}
              style={[styles.chip, status === value && styles.chipActive]}
              onPress={() => setStatus(value)}
            >
              <Text style={status === value ? styles.chipTextActive : styles.chipText}>
                {STATUS_LABELS[value]}
              </Text>
            </Pressable>
          ))}
        </View>

        <Text style={styles.label}>Assignee</Text>
        <TextInput
          style={styles.input}
          value={assignee}
          onChangeText={setAssignee}
          placeholder="Assignee"
        />

        {error ? <Text style={styles.error}>{error}</Text> : null}

        <View style={styles.actions}>
          <Pressable style={styles.cancel} onPress={onClose} disabled={isSaving}>
            <Text>Cancel</Text>
          </Pressable>
          <Pressable style={styles.save} onPress={submit} disabled={isSaving}>
            <Text style={styles.saveText}>{isSaving ? 'Saving...' : 'Save'}</Text>
          </Pressable>
        </View>
      </ScrollView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: { padding: 20, paddingTop: 48 },
  heading: { fontSize: 20, fontWeight: '700', marginBottom: 16 },
  label: { fontSize: 13, color: '#555', marginTop: 12, marginBottom: 4 },
  input: { borderWidth: 1, borderColor: '#ccc', borderRadius: 6, padding: 10, fontSize: 15 },
  multiline: { minHeight: 80, textAlignVertical: 'top' },
  chips: { flexDirection: 'row', gap: 8 },
  chip: { borderWidth: 1, borderColor: '#ccc', borderRadius: 16, paddingVertical: 6, paddingHorizontal: 12 },
  chipActive: { backgroundColor: '#2563eb', borderColor: '#2563eb' },
  chipText: { color: '#333' },
  chipTextActive: { color: '#fff' },
  error: { color: '#b91c1c', marginTop: 12 },
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: 12, marginTop: 24 },
  cancel: { paddingVertical: 10, paddingHorizontal: 16 },
  save: { backgroundColor: '#2563eb', borderRadius: 6, paddingVertical: 10, paddingHorizontal: 20 },
  saveText: { color: '#fff', fontWeight: '600' },
});
