import type { User } from '~/types/items';

/** The plan of the account (pro: the Pro mode's switch), and has(feature) to show a feature only where it is available. ready: the account is loaded, has() is final. */
export function usePro() {
  const api = useApi();
  const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));
  const pro = computed(() => me.value?.pro ?? false);
  const plan = computed(() => (me.value ? planOf(me.value) : 'base'));
  const has = (feature: Feature) => unlocked(features[feature], plan.value);
  const ready = computed(() => !!me.value);
  return { pro, plan, has, ready };
}
