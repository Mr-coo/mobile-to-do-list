import { useEffect, useState } from 'react';
import { StyleSheet, TextInput } from 'react-native';

interface Props {
  onSearch: (keyword: string) => void;
  delayMs?: number;
}

export function SearchBar({ onSearch, delayMs = 300 }: Props) {
  const [text, setText] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => onSearch(text.trim()), delayMs);
    return () => clearTimeout(timer);
  }, [text, delayMs, onSearch]);

  return (
    <TextInput
      testID="search-input"
      style={styles.input}
      value={text}
      onChangeText={setText}
      placeholder="Search tasks"
      autoCorrect={false}
    />
  );
}

const styles = StyleSheet.create({
  input: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#d4d4d4',
    borderRadius: 6,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
});
