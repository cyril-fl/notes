<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

definePageMeta({ feature: 'graph' });

const { icons } = useIcons();

const route = useRoute();
const api = useApi();

const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
</script>

<template>
  <div class="flex h-full flex-col p-8">
    <PageHeader :title="$t('common.graph')" :subtitle="$t('graph.subtitle')" />
    <template v-if="graph">
      <GraphView
        v-if="graph.notes.length"
        class="min-h-0 flex-1"
        :graph="graph"
        :selected="route.query.note as string | undefined"
      />
      <UEmpty
        v-else
        :icon="icons.graph"
        :title="$t('graph.empty')"
        :description="$t('graph.emptyDescription')"
        :actions="[{ label: $t('common.newNote'), icon: icons.noteAdd, to: newNotePath() }]"
      />
    </template>
  </div>
</template>
