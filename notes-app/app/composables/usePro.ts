/** The plan of the account, and has(feature) to show a feature only where it is available. ready: the account is loaded, has() is final. */
export function usePro() {
  const { data: me } = useMe();
  const plan = computed(() => (me.value ? planOf(me.value) : 'base'));
  const has = (feature: Feature) => unlocked(features[feature], plan.value);
  const ready = computed(() => !!me.value);
  return { plan, has, ready };
}
