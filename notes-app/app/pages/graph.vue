<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const route = useRoute();
const api = useApi();

const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
</script>

<template>
  <div class="flex h-full flex-col p-8">
    <PageHeader title="Graphe" subtitle="Glissez pour vous déplacer, molette ou pincement pour zoomer" />
    <template v-if="graph">
      <GraphView v-if="graph.notes.length" class="min-h-0 flex-1" :graph="graph" :selected="route.query.note as string | undefined" />
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
