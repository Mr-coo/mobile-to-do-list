import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { TasksScreen } from './src/screens/TasksScreen';

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <SafeAreaProvider>
        <TasksScreen />
      </SafeAreaProvider>
    </QueryClientProvider>
  );
}
