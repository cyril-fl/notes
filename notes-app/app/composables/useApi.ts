import type { ToastProps } from '@nuxt/ui';
import type { FetchError, FetchOptions } from 'ofetch';

type ApiOptions = FetchOptions<'json'> & {
  /** What the toast says on a 409, the API's messages being in English: one text, or one per part of the API's message. */
  conflict?: string | Record<string, string>;
  /** The buttons of the toast on a 409. */
  conflictActions?: ToastProps['actions'];
  /** What the toast says on a 400, by part of the API's message. */
  invalid?: Record<string, string>;
  /** What the toast says on a 413, the limit of this call; a general text otherwise. */
  tooLarge?: string;
  /** The page shows the error itself: always, or for these statuses. */
  quiet?: boolean | number[];
};

function errorMessage(
  error: FetchError,
  t: (key: string) => string,
  { conflict, invalid, tooLarge }: ApiOptions,
): string {
  if (!error.status) return t('api.down');
  if (error.status === 413) return tooLarge ?? t('api.tooLarge');
  const message: string = error.data?.error ?? t('api.failed');
  const texts = { 400: invalid, 409: conflict }[error.status];
  if (!texts) return message;
  if (typeof texts === 'string') return texts;
  return Object.entries(texts).find(([part]) => message.includes(part))?.[1] ?? message;
}

/** Calls the Go API and tells the user when it fails; without a session, opens the login page. */
export function useApi() {
  const toast = useToast();
  const { t } = useI18n();
  const route = useRoute();
  const baseURL = useRuntimeConfig().public.apiBase;

  return async function api<T>(
    path: string,
    { conflict, conflictActions, invalid, tooLarge, quiet, ...options }: ApiOptions = {},
  ): Promise<T> {
    try {
      // include: in development the API is on another port, and its session cookie must go along.
      return await $fetch<T>(path, { baseURL, credentials: 'include', ...options } as object);
    } catch (error) {
      const status = (error as FetchError).status;
      if (status === 401 && route.path !== '/login') {
        await navigateTo({ path: '/login', query: { redirect: route.fullPath } });
      } else if (quiet !== true && !(Array.isArray(quiet) && quiet.includes(status ?? 0))) {
        toast.add({
          title: errorMessage(error as FetchError, t, { conflict, invalid, tooLarge }),
          color: 'error',
          ...(status === 409 && conflictActions && { actions: conflictActions, duration: 0 }),
        });
      }
      throw error;
    }
  };
}
