import type { User } from '~/types/items';

/** The account, loaded once for the whole app: every row of a folder asks for it, through usePro. */
export function useMe() {
  const api = useApi();
  return useLazyAsyncData('me', () => api<User>('/api/me'), {
    // Nuxt's default cancels the pending GET and sends another one for each new caller: they share it instead.
    dedupe: 'defer',
    // Lazy, each component fetches again once mounted: it takes the account already loaded. A refresh still fetches.
    getCachedData: (key, nuxtApp, { cause }) => (cause === 'initial' ? nuxtApp.payload.data[key] : undefined),
  });
}
