import { render } from '@testing-library/react-native';
import { TaskList } from '../TaskList';
import { Task } from '../../types';

const tasks: Task[] = [
  {
    id: 1,
    title: 'Write docs',
    description: 'README and API notes',
    status: 'todo',
    assignee: 'Sam',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    title: 'Fix login bug',
    description: null,
    status: 'in_progress',
    assignee: null,
    created_at: '2026-01-02T00:00:00Z',
    updated_at: '2026-01-02T00:00:00Z',
  },
];

describe('TaskList', () => {
  it('renders the title of each task', async () => {
    const { getByText } = await render(<TaskList tasks={tasks} isLoading={false} />);
    expect(getByText('Write docs')).toBeTruthy();
    expect(getByText('Fix login bug')).toBeTruthy();
  });

  it('shows a spinner while loading', async () => {
    const { getByTestId } = await render(<TaskList tasks={[]} isLoading />);
    expect(getByTestId('task-list-loading')).toBeTruthy();
  });

  it('shows an empty message when there are no tasks', async () => {
    const { getByText } = await render(<TaskList tasks={[]} isLoading={false} />);
    expect(getByText('No tasks found')).toBeTruthy();
  });
});
