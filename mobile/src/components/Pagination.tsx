import { Pressable, StyleSheet, Text, View } from 'react-native';

interface Props {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
  disabled?: boolean;
}

export function Pagination({ page, totalPages, onChange, disabled }: Props) {
  if (totalPages <= 1) return null;

  const prevDisabled = disabled || page <= 1;
  const nextDisabled = disabled || page >= totalPages;

  return (
    <View style={styles.row}>
      <Pressable
        accessibilityRole="button"
        accessibilityState={{ disabled: prevDisabled }}
        disabled={prevDisabled}
        onPress={() => onChange(page - 1)}
      >
        <Text style={[styles.link, prevDisabled && styles.linkDisabled]}>‹ Prev</Text>
      </Pressable>
      <Text>{`Page ${page} of ${totalPages}`}</Text>
      <Pressable
        accessibilityRole="button"
        accessibilityState={{ disabled: nextDisabled }}
        disabled={nextDisabled}
        onPress={() => onChange(page + 1)}
      >
        <Text style={[styles.link, nextDisabled && styles.linkDisabled]}>Next ›</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', padding: 16 },
  link: { color: '#2563eb', fontWeight: '600' },
  linkDisabled: { color: '#aaa' },
});
