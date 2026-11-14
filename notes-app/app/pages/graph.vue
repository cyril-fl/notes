<script setup lang="ts">
import type { NotesGraph, User } from '~/types/items';

definePageMeta({ feature: 'graph' });

const { icons } = useIcons();

const route = useRoute();
const api = useApi();

// From a note: its neighborhood, as the mini graph of its panel. Otherwise every note.
const noteId = computed(() => (typeof route.query.note === 'string' ? route.query.note : undefined));
const { data: neighbors } = useNeighbors(noteId);
const { data: all } = useLazyAsyncData(
  'graph',
  () => (noteId.value ? Promise.resolve(null) : api<NotesGraph>('/api/graph/notes')),
  { watch: [noteId] },
);
const me = useNuxtData<User>('me').data;
// The tags on most notes, left out on demand: they say little of each note and pull them all together. Remembered on this device.
const hideWide = useLocalStorage('graph-hide-wide', false);
const narrow = computed(() => all.value && withoutWideTags(all.value));
const graph = computed(() => {
  if (noteId.value) return neighbors.value && neighborGraph(neighbors.value);
  return hideWide.value ? narrow.value?.graph : all.value;
});

const selectedNotePath = computed(() => {
  const path = neighbors.value?.note.path;
  return path?.slice(0, path.lastIndexOf('/')) || '/';
});
</script>

<template>
  <div class="flex h-full flex-col p-8">
    <PageHeader
      :title="$t('common.graph')"
      :subtitle="neighbors && noteId ? displayTitle(neighbors.note.item.title, $t('untitled')) : undefined"
    >
      <USwitch
        v-if="!noteId && narrow?.wide.length"
        v-model="hideWide"
        :label="$t('graph.hideWide', { count: narrow.wide.length, share: WIDE_TAG_SHARE * 100 })"
        :title="narrow.wide.map((tag) => `#${tag}`).join(' ')"
        size="sm"
      />
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
      <GraphView
        v-if="graph.notes.length"
        class="min-h-0 flex-1"
        :graph="graph"
        :selected="noteId"
        :cache="noteId ? undefined : `graph:${me?.id}:${hideWide ? 'narrow' : 'all'}`"
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
