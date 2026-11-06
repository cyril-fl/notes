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
      <UEmpty
        v-else
        icon="i-lucide-waypoints"
        title="Aucune note pour le moment"
        description="Ajoutez des #tags à vos notes : celles qui en partagent un sont reliées ici."
        :actions="[{ label: 'Nouvelle note', icon: 'i-lucide-square-pen', to: newNotePath() }]"
      />
    </template>
  </div>
</template>
