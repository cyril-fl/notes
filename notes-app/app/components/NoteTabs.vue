<script setup lang="ts">
import type { NotesGraph } from '~/types/items';

const { icons } = useIcons();
const { has } = usePro();
const { t } = useI18n();
const api = useApi();
const route = useRoute();
const tabs = useTabs();
// The focus mode leaves only the note on screen: the tabs go with the sidebar.
const focus = useState('focus', () => false);

// The titles come from the notes graph, which the editor loads under the same key: no request of its own.
const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
const titles = computed(() => new Map(graph.value?.notes.map((note) => [note.id, note.title])));
const label = (id: string) => displayTitle(titles.value.get(id), t('untitled'));

// A new note takes its tab once created: the editor then redirects to its address.
const open = computed(() => (route.params.id === 'new' ? undefined : (route.params.id as string | undefined)));
watch(open, (id) => id && !tabs.value.includes(id) && tabs.value.push(id), { immediate: true });

// A note removed anywhere, a folder's notes with it, leaves the graph when the views read again: its tab follows.
// The open one stays: a note just created is not in the graph yet.
watch(graph, (loaded) => {
  if (!loaded) return;
  tabs.value = tabs.value.filter((id) => titles.value.has(id) || id === open.value);
});

async function close(id: string) {
  const next = tabAfterClose(tabs.value, id);
  tabs.value = tabs.value.filter((tab) => tab !== id);
  if (id === open.value) await navigateTo(next ? `/notes/${next}` : '/');
}

const shown = computed(() => has('tabs') && tabs.value.length > 0 && !(focus.value && has('focusMode')));
</script>

<template>
  <nav
    v-if="shown"
    :aria-label="$t('note.tabs')"
    class="flex shrink-0 gap-1 overflow-x-auto border-b border-default px-2 py-1"
  >
    <div
      v-for="id in tabs"
      :key="id"
      class="flex shrink-0 items-center rounded-md pr-1"
      :class="id === open ? 'bg-elevated' : 'hover:bg-elevated/50'"
    >
      <NuxtLink
        :to="`/notes/${id}`"
        class="max-w-40 truncate py-1 pl-2 text-sm"
        :class="id === open ? 'text-default' : 'text-muted'"
      >
        {{ label(id) }}
      </NuxtLink>
      <UButton
        :icon="icons.close"
        color="neutral"
        variant="ghost"
        size="xs"
        :aria-label="$t('note.closeTab', { title: label(id) })"
        @click="close(id)"
      />
    </div>
  </nav>
</template>
