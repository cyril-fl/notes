import type { FetchError, FetchOptions } from 'ofetch';

type ApiOptions = FetchOptions<'json'> & {
  /** What the toast says on a 409, the API's messages being in English: one text, or one per part of the API's message. */
  conflict?: string | Record<string, string>;
  /** What the toast says on a 413, the limit of this call; a general text otherwise. */
  tooLarge?: string;
  /** The page shows the error itself. */
  quiet?: boolean;
};

function errorMessage(error: FetchError, t: (key: string) => string, { conflict, tooLarge }: ApiOptions): string {
  if (!error.status) return t('api.down');
  if (error.status === 413) return tooLarge ?? t('api.tooLarge');
  const message: string = error.data?.error ?? t('api.failed');
  if (error.status !== 409 || !conflict) return message;
  if (typeof conflict === 'string') return conflict;
  return Object.entries(conflict).find(([part]) => message.includes(part))?.[1] ?? message;
}

/** Calls the Go API and tells the user when it fails; without a session, opens the login page. */
export function useApi() {
  const toast = useToast();
  const { t } = useI18n();
  const route = useRoute();
  const baseURL = useRuntimeConfig().public.apiBase;

  return async function api<T>(path: string, { conflict, tooLarge, quiet, ...options }: ApiOptions = {}): Promise<T> {
    try {
      // include: in development the API is on another port, and its session cookie must go along.
      return await $fetch<T>(path, { baseURL, credentials: 'include', ...options } as object);
    } catch (error) {
      if ((error as FetchError).status === 401 && route.path !== '/login') {
        await navigateTo({ path: '/login', query: { redirect: route.fullPath } });
      } else if (!quiet) {
        toast.add({ title: errorMessage(error as FetchError, t, { conflict, tooLarge }), color: 'error' });
      }
      throw error;
    }
  };
}
