import type { FetchError, FetchOptions } from 'ofetch';

type ApiOptions = FetchOptions<'json'> & {
  /** What the toast says on a 409, the API's messages being in English: one text, or one per part of the API's message. */
  conflict?: string | Record<string, string>;
  /** The page shows the error itself. */
  quiet?: boolean;
};

function errorMessage(error: FetchError, conflict?: ApiOptions['conflict']): string {
  if (!error.status) return "L'API ne répond pas.";
  const message: string = error.data?.error ?? 'Une erreur est survenue.';
  if (error.status !== 409 || !conflict) return message;
  if (typeof conflict === 'string') return conflict;
  return Object.entries(conflict).find(([part]) => message.includes(part))?.[1] ?? message;
}

/** Calls the Go API and tells the user when it fails. */
export function useApi() {
  const toast = useToast();
  const baseURL = useRuntimeConfig().public.apiBase;

  return async function api<T>(path: string, { conflict, quiet, ...options }: ApiOptions = {}): Promise<T> {
    try {
      return await $fetch<T>(path, { baseURL, ...options } as object);
    } catch (error) {
      if (!quiet) toast.add({ title: errorMessage(error as FetchError, conflict), color: 'error' });
      throw error;
    }
  };
}
