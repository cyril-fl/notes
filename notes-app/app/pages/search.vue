<script setup lang="ts">
import type { Match } from '~/types/items';

const route = useRoute();
const api = useApi();

const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''));

const { data: matches, status } = useLazyAsyncData(
  'search',
  () => (q.value.trim() ? api<Match[]>('/api/search', { query: { q: q.value } }) : Promise.resolve([])),
  { watch: [q], default: () => [] as Match[] }
);
</script>

<template>
  <div class="mx-auto max-w-3xl p-8">
    <PageHeader title="Recherche" :subtitle="q" />
    <NoteList v-if="matches.length" :notes="matches" />
    <p v-else-if="status !== 'pending'" class="py-16 text-center text-sm text-muted">Aucun résultat.</p>
  </div>
</template>
