<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

definePageMeta({ feature: 'graph' });

const { icons } = useIcons();

const route = useRoute();
const api = useApi();

const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));

const selectedNotePath = computed(() => {
  const noteId = typeof route.query.note === 'string' ? route.query.note : undefined;
  const path = graph.value?.notes.find((note) => note.id === noteId)?.path;
  return path?.slice(0, path.lastIndexOf('/')) || '/';
});
</script>

<template>
  <div class="flex h-full flex-col p-8">
    <PageHeader :title="$t('common.graph')" />
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
    <footer
      v-if="route.query.note"
      class="sticky bottom-0 z-10 -mx-8 -mb-8 mt-auto border-t border-default bg-default px-8"
    >
      <div class="flex h-9 items-center text-xs text-muted">
        <PathNav :path="selectedNotePath" />
      </div>
    </footer>
  </div>
</template>
