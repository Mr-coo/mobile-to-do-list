import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { ApiError } from '../../api/client';
import { listTasks, updateTask } from '../../api/tasks';
import { ListResponse, Task } from '../../types';
import { TasksScreen } from '../TasksScreen';

jest.mock('../../api/tasks');
jest.mock('../../components/SearchBar', () => ({ SearchBar: () => null }));

const mockListTasks = jest.mocked(listTasks);
const mockUpdateTask = jest.mocked(updateTask);

const insets = { frame: { x: 0, y: 0, width: 360, height: 640 }, insets: { top: 0, left: 0, right: 0, bottom: 0 } };

function task(title: string): Task {
  return {
    id: 1,
    title,
    description: null,
    status: 'todo',
    assignee: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  };
}

function page(tasks: Task[]): ListResponse {
  return { data: tasks, meta: { page: 1, limit: 10, total: tasks.length, total_pages: 1 } };
}

async function renderScreen() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity }, mutations: { retry: false, gcTime: Infinity } },
  });
  await render(
    <SafeAreaProvider initialMetrics={insets}>
      <QueryClientProvider client={client}>
        <TasksScreen />
      </QueryClientProvider>
    </SafeAreaProvider>,
  );
}

beforeEach(() => {
  jest.resetAllMocks();
});

describe('TasksScreen', () => {
  it('refreshes the list after saving an edit', async () => {
    mockListTasks
      .mockResolvedValueOnce(page([task('Buy milk')]))
      .mockResolvedValue(page([task('Buy oat milk')]));
    mockUpdateTask.mockResolvedValue({ data: task('Buy oat milk') });
    await renderScreen();

    await fireEvent.press(await screen.findByText('Buy milk'));
    expect(await screen.findByText('Edit task')).toBeTruthy();
    await fireEvent.changeText(screen.getByDisplayValue('Buy milk'), 'Buy oat milk');
    await fireEvent.press(screen.getByText('Save'));

    expect(await screen.findByText('Buy oat milk')).toBeTruthy();
    expect(mockUpdateTask).toHaveBeenCalledWith(1, expect.objectContaining({ title: 'Buy oat milk' }));
    expect(mockListTasks).toHaveBeenCalledTimes(2);
    await waitFor(() => expect(screen.queryByText('Edit task')).toBeNull());
  });

  it('shows duplicate title error and keeps the modal open', async () => {
    mockListTasks.mockResolvedValue(page([task('Buy milk')]));
    mockUpdateTask.mockRejectedValue(new ApiError(409, 'DUPLICATE_TITLE', 'duplicate'));
    await renderScreen();

    await fireEvent.press(await screen.findByText('Buy milk'));
    await screen.findByText('Edit task');
    await fireEvent.press(screen.getByText('Save'));

    expect(await screen.findByText('A task with this title already exists')).toBeTruthy();
    expect(screen.getByText('Edit task')).toBeTruthy();
  });

  it('shows validation details under the field', async () => {
    mockListTasks.mockResolvedValue(page([task('Buy milk')]));
    mockUpdateTask.mockRejectedValue(
      new ApiError(400, 'VALIDATION_ERROR', 'Invalid input', { title: 'title must be at most 255 characters' }),
    );
    await renderScreen();

    await fireEvent.press(await screen.findByText('Buy milk'));
    await screen.findByText('Edit task');
    await fireEvent.press(screen.getByText('Save'));

    expect(await screen.findByText('title must be at most 255 characters')).toBeTruthy();
  });

  it('shows a generic message for other errors', async () => {
    mockListTasks.mockResolvedValue(page([task('Buy milk')]));
    mockUpdateTask.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'boom'));
    await renderScreen();

    await fireEvent.press(await screen.findByText('Buy milk'));
    await screen.findByText('Edit task');
    await fireEvent.press(screen.getByText('Save'));

    expect(await screen.findByText('Something went wrong, try again')).toBeTruthy();
  });

  it('shows only the error state when loading fails', async () => {
    mockListTasks.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'boom'));
    await renderScreen();

    expect(await screen.findByText("Couldn't load tasks")).toBeTruthy();
    expect(screen.queryByText('No tasks found')).toBeNull();

    mockListTasks.mockResolvedValue(page([task('Buy milk')]));
    await fireEvent.press(screen.getByText('Retry'));

    expect(await screen.findByText('Buy milk')).toBeTruthy();
    expect(mockListTasks).toHaveBeenCalledTimes(2);
  });

  it('only offers Delete when editing', async () => {
    mockListTasks.mockResolvedValue(page([task('Buy milk')]));
    await renderScreen();

    await fireEvent.press(await screen.findByText('+ New task'));
    await screen.findByText('New task');
    expect(screen.queryByText('Delete')).toBeNull();
  });
});
