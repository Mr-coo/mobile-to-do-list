import { act, fireEvent, render, screen } from '@testing-library/react-native';
import { SearchBar } from '../SearchBar';

describe('SearchBar', () => {
  beforeEach(() => jest.useFakeTimers());
  afterEach(() => jest.useRealTimers());

  it('calls onSearch once with the trimmed value after 300ms', async () => {
    const onSearch = jest.fn();
    await render(<SearchBar onSearch={onSearch} />);
    await fireEvent.changeText(screen.getByTestId('search-input'), ' milk ');

    await act(async () => {
      jest.advanceTimersByTime(299);
    });
    expect(onSearch).not.toHaveBeenCalled();

    await act(async () => {
      jest.advanceTimersByTime(1);
    });
    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch).toHaveBeenCalledWith('milk');
  });

  it('debounces rapid typing into a single call', async () => {
    const onSearch = jest.fn();
    await render(<SearchBar onSearch={onSearch} />);
    const input = screen.getByTestId('search-input');

    await fireEvent.changeText(input, 'mi');
    await act(async () => {
      jest.advanceTimersByTime(100);
    });
    await fireEvent.changeText(input, 'milk');
    await act(async () => {
      jest.advanceTimersByTime(300);
    });

    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch).toHaveBeenCalledWith('milk');
  });
});
