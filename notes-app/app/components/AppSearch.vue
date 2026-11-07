<script setup lang="ts">
import type { CommandPaletteGroup, CommandPaletteItem } from '@nuxt/ui';
import type { Match, Recent } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

const api = useApi();
const route = useRoute();
const { createFolder } = useItems();

const open = useState('search-open', () => false);
const term = ref('');
const query = refDebounced(term, 200);

const { data: matches, status } = useLazyAsyncData(
  'palette-search',
  () =>
    query.value.trim() ? api<Match[]>('/api/search', { query: { q: query.value }, quiet: true }) : Promise.resolve([]),
  { watch: [query], default: () => [] as Match[] },
);
const { data: recent } = useLazyAsyncData('recent', () => api<Recent[]>('/api/recent', { quiet: true }), {
  default: () => [] as Recent[],
});

const toNote = (note: { id: string; title: string; path: string }): CommandPaletteItem => ({
  label: displayTitle(note.title, t('untitled')),
  suffix: note.path,
  icon: icons.note,
  to: `/notes/${note.id}`,
});

// The API searches the text: the palette shows what it found as is, not filtered again on titles.
const groups = computed<CommandPaletteGroup[]>(() => {
  const searching = term.value.trim() !== '';
  return [
    searching
      ? { id: 'notes', label: t('common.notes'), items: matches.value.map(toNote), ignoreFilter: true }
      : { id: 'recent', label: t('sidebar.recent'), items: recent.value.slice(0, 8).map(toNote) },
    {
      id: 'actions',
      label: t('search.actions'),
      items: [
        { label: t('common.newNote'), icon: icons.noteAdd, kbds: ['N'], to: newNotePath(currentFolder(route)) },
        {
          label: t('common.newFolder'),
          icon: icons.folderAdd,
          kbds: ['shift', 'N'],
          onSelect: () => createFolder(currentFolder(route)),
        },
        { label: t('common.graph'), icon: icons.graph, to: '/graph' },
        { label: t('common.trash'), icon: icons.trash, to: '/trash' },
        { label: t('common.settings'), icon: icons.settings, to: '/settings' },
        ...(searching
          ? [
              {
                label: t('search.allResults', { term: term.value }),
                icon: icons.search,
                to: { path: '/search', query: { q: term.value } },
              },
            ]
          : []),
      ],
    },
  ];
});
</script>

<template>
  <UDashboardSearch
    v-model:open="open"
    v-model:search-term="term"
    :groups="groups"
    :loading="status === 'pending'"
    :placeholder="$t('search.placeholder')"
  />
</template>
