export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080/api';

export class ApiError extends Error {
  status: number;
  code: string;
  details?: Record<string, string>;

  constructor(status: number, code: string, message: string, details?: Record<string, string>) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  });

  let body: any = null;
  try {
    body = await response.json();
  } catch {
    body = null;
  }

  if (!response.ok) {
    const error = body?.error;
    if (error && typeof error === 'object') {
      throw new ApiError(response.status, error.code ?? 'UNKNOWN', error.message ?? 'Request failed', error.details);
    }
    throw new ApiError(response.status, 'UNKNOWN', typeof error === 'string' ? error : 'Request failed');
  }

  return body as T;
}
