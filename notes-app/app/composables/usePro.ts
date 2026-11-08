import type { User } from '~/types/items';

/** The Pro mode of the account, and has(feature) to show a feature only where it is available. ready: the account is loaded, has() is final. */
export function usePro() {
  const api = useApi();
  const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));
  const pro = computed(() => me.value?.pro ?? false);
  const has = (feature: Feature) => unlocked(features[feature], pro.value);
  const ready = computed(() => !!me.value);
  return { pro, has, ready };
}
