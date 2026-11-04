<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const route = useRoute();
const api = useApi();

const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
</script>

<template>
  <div class="mx-auto max-w-5xl p-8">
    <PageHeader title="Graphe" />
    <template v-if="graph">
      <GraphView v-if="graph.notes.length" :graph="graph" :selected="route.query.note as string | undefined" />
      <p v-else class="py-16 text-center text-sm text-muted">Aucune note pour le moment.</p>
    </template>
  </div>
</template>
