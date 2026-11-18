import { fireEvent, render, screen } from '@testing-library/react-native';
import { Pagination } from '../Pagination';

describe('Pagination', () => {
  it('disables Prev on the first page and goes to the next page', async () => {
    const onChange = jest.fn();
    await render(<Pagination page={1} totalPages={3} onChange={onChange} />);

    expect(screen.getByText('Page 1 of 3')).toBeTruthy();
    await fireEvent.press(screen.getByText('‹ Prev'));
    expect(onChange).not.toHaveBeenCalled();

    await fireEvent.press(screen.getByText('Next ›'));
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it('disables Next on the last page', async () => {
    const onChange = jest.fn();
    await render(<Pagination page={3} totalPages={3} onChange={onChange} />);

    await fireEvent.press(screen.getByText('Next ›'));
    expect(onChange).not.toHaveBeenCalled();

    await fireEvent.press(screen.getByText('‹ Prev'));
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it('renders nothing when there is a single page', async () => {
    await render(<Pagination page={1} totalPages={1} onChange={jest.fn()} />);
    expect(screen.queryByText('Page 1 of 1')).toBeNull();
  });
});
