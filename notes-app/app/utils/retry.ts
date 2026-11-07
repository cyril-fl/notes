import type { FetchError } from 'ofetch';

/** How long the API asks to wait after wrong passwords, from its Retry-After seconds: "30 s", "5 min", "1 h". */
export function retryWait(error: FetchError): string {
  const seconds = Number(error.response?.headers.get('Retry-After')) || 0;
  if (seconds < 60) return `${seconds} s`;
  if (seconds < 3600) return `${Math.ceil(seconds / 60)} min`;
  return `${Math.ceil(seconds / 3600)} h`;
}
