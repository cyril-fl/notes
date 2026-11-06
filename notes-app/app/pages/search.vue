<script setup lang="ts">
import type { Match } from '~/types/items';

const route = useRoute();
const api = useApi();
const view = useView();

const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''));

const { data: matches, status } = useLazyAsyncData(
  'search',
  () => (q.value.trim() ? api<Match[]>('/api/search', { query: { q: q.value } }) : Promise.resolve([])),
  { watch: [q], default: () => [] as Match[] }
);
</script>

<template>
  <div class="mx-auto max-w-3xl p-8">
    <PageHeader title="Recherche" :subtitle="q">
      <ViewToggle />
    </PageHeader>
    <NoteList v-if="matches.length" :notes="matches" :grid="view === 'grid'" />
    <UEmpty
      v-else-if="status !== 'pending'"
      icon="i-lucide-search-x"
      :title="q.trim() ? 'Aucun résultat' : 'Rechercher dans les notes'"
      :description="q.trim() ? `Aucune note ne contient « ${q} ».` : 'Tapez un mot dans la barre de recherche.'"
    />
  </div>
</template>
