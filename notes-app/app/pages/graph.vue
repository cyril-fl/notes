<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

definePageMeta({ feature: 'graph' });

const { icons } = useIcons();
const appConfig = useAppConfig();

const route = useRoute();
const api = useApi();

// From a note: its neighborhood, as the mini graph of its panel. Otherwise every note. Both laid out by the server.
const noteId = computed(() => (typeof route.query.note === 'string' ? route.query.note : undefined));
const { data: near } = useNoteGraph(noteId);
const { data: all } = useLazyAsyncData(
  'graph',
  () => (noteId.value ? Promise.resolve(null) : api<NotesGraph>('/api/graph/notes')),
  { watch: [noteId] },
);
const graph = computed(() => (noteId.value ? near.value : all.value));
// The neighborhood starts with its note.
const center = computed(() => (noteId.value ? near.value?.notes[0] : undefined));

const selectedNotePath = computed(() => {
  const path = center.value?.path;
  return path?.slice(0, path.lastIndexOf('/')) || '/';
});
</script>

<template>
  <div class="flex h-full flex-col p-8">
    <PageHeader :title="$t('common.graph')" :subtitle="center ? displayTitle(center.title, $t('untitled')) : undefined">
      <UButton
        v-if="noteId"
        :label="$t('graph.showAll')"
        :icon="icons.graph"
        to="/graph"
        color="neutral"
        variant="ghost"
        size="sm"
      />
    </PageHeader>
    <template v-if="graph">
      <GraphView v-if="graph.notes.length" class="min-h-0 flex-1" :graph="graph" :selected="noteId" />
      <UEmpty
        v-else
        :icon="icons.graph"
        :title="$t('graph.empty')"
        :description="$t('graph.emptyDescription')"
        :actions="[{ label: $t('common.newNote'), icon: icons.noteAdd, to: newNotePath() }]"
      />
    </template>
    <!-- The first layout of a big account takes the server a few seconds. -->
    <div v-else class="flex flex-1 items-center justify-center" role="status">
      <UIcon
        :name="appConfig.ui.icons.loading"
        class="size-5 animate-spin text-muted"
        :aria-label="$t('graph.loading')"
      />
    </div>
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
