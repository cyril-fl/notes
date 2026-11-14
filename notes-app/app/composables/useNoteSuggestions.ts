import type { TagCount } from '~/types/items';

/** The menus of the note editor that insert text: the tags after #, the snippets after ;; */
export function useNoteSuggestions(): EditorModule {
  const { icons } = useIcons();
  const api = useApi();

  // The snippets offered after ;; : picked, their markdown goes in at the cursor.
  const { data: me } = useMe();
  const snippetItems = computed(() =>
    (me.value?.snippets ?? []).map((snippet) => ({
      kind: 'snippet',
      label: snippet.name,
      icon: icons.snippet,
      content: snippet.content,
    })),
  );

  // The existing tags offered after #: the same data as the sidebar, refreshed when a save changes the tags.
  const { data: allTags } = useLazyAsyncData('tags', () => api<TagCount[]>('/api/graph'), {
    default: () => [] as TagCount[],
  });
  const tagItems = computed(() => allTags.value.map((tag) => ({ kind: 'tag', label: tag.value, icon: icons.tag })));

  return {
    handlers: {
      tag: command((editor, item: { label: string }) => editor.chain().focus().insertContent(`#${item.label} `)),
      snippet: command((editor, item: { content: string }) =>
        editor.chain().focus().insertContent(item.content, { contentType: 'markdown' }),
      ),
    },
    menus: [
      { char: ';;', pluginKey: 'snippetMenu', items: snippetItems, suggestion: { allowSpaces: true } },
      { char: '#', pluginKey: 'tagMenu', items: tagItems },
    ],
  };
}
