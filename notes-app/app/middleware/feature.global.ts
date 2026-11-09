import type { User } from '~/types/items';

// A page of a feature the account lacks, definePageMeta({ feature }), sends home: also when its URL is typed by hand.
export default defineNuxtRouteMiddleware(async (to) => {
  const feature = to.meta.feature as Feature | undefined;
  if (!feature || features[feature] === 'base') return;
  const me = useNuxtData<User>('me').data.value ?? (await useApi()<User>('/api/me'));
  if (!unlocked(features[feature], planOf(me))) return navigateTo('/');
});
