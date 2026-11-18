import { ApiError, request } from '../client';

function mockFetch(status: number, body?: unknown) {
  const json = jest.fn().mockResolvedValue(body);
  globalThis.fetch = jest.fn().mockResolvedValue({ ok: status < 400, status, json }) as unknown as typeof fetch;
  return json;
}

describe('request', () => {
  it('throws an ApiError with status, code, message and details for the standard error body', async () => {
    mockFetch(400, {
      error: { code: 'VALIDATION_ERROR', message: 'Invalid input', details: { title: 'title is required' } },
    });

    const error = await request('/tasks').catch((e: unknown) => e) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(400);
    expect(error.code).toBe('VALIDATION_ERROR');
    expect(error.message).toBe('Invalid input');
    expect(error.details).toEqual({ title: 'title is required' });
  });

  it('falls back to code UNKNOWN when the error is a plain string', async () => {
    mockFetch(500, { error: 'some string' });

    const error = await request('/tasks').catch((e: unknown) => e) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.code).toBe('UNKNOWN');
    expect(error.message).toBe('some string');
  });

  it('resolves undefined for 204 without parsing JSON', async () => {
    const json = mockFetch(204);

    await expect(request('/tasks/1', { method: 'DELETE' })).resolves.toBeUndefined();
    expect(json).not.toHaveBeenCalled();
  });
});
