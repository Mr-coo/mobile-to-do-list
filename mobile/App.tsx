import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TasksScreen } from './src/screens/TasksScreen';

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TasksScreen />
    </QueryClientProvider>
  );
}
