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

interface ErrorShape {
  code?: string;
  message?: string;
  details?: Record<string, string>;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function toApiError(status: number, body: unknown): ApiError {
  const error = isRecord(body) ? body.error : undefined;
  if (isRecord(error)) {
    const { code, message, details } = error as ErrorShape;
    return new ApiError(status, code ?? 'UNKNOWN', message ?? 'Request failed', details);
  }
  return new ApiError(status, 'UNKNOWN', typeof error === 'string' ? error : 'Request failed');
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  });

  if (response.status === 204) {
    return undefined as T;
  }

  let body: unknown = null;
  try {
    body = await response.json();
  } catch {
    body = null;
  }

  if (!response.ok) {
    throw toApiError(response.status, body);
  }

  return body as T;
}
