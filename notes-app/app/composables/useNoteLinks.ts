import type { Editor } from '@tiptap/vue-3';
import type { NotesGraph } from '~/types/items';
import type { LinkSize } from '~/utils/linkCard';
import LinkCard from '~/components/LinkCard.vue';

const ALLOWED_LINKS = ['http', 'https', 'mailto'];

/** The links of the note editor: web links and their cards, [[links]] to the other notes. id: the note, left out of the [[ menu. */
export function useNoteLinks(id: Ref<string | undefined>): EditorModule {
  const { icons } = useIcons();
  const { t } = useI18n();
  const api = useApi();
  const { prompt } = useDialogs();
  const toast = useToast();

  async function askLink(editor: Editor) {
    const current = editor.getAttributes('link').href as string | undefined;
    if (current) return editor.chain().focus().unsetLink().run();

    const href = await prompt({
      title: t('note.addLink'),
      placeholder: t('note.linkPlaceholder'),
      action: t('common.add'),
    });
    if (!href) return;
    const { from, to } = editor.state.selection;
    editor
      .chain()
      .focus()
      .insertContent(linkContent(toHref(href), editor.state.doc.textBetween(from, to)))
      .run();
  }

  // A link swapped in place to another size, the same href and text.
  function setLinkSize(editor: Editor, size: LinkSize) {
    const card = editor.isActive('linkCard') && editor.getAttributes('linkCard');
    if (card)
      return editor
        .chain()
        .focus()
        .insertContent(linkContent(card.href, card.text, size));
    if (size === 'hyperlink') return editor.chain();
    const href = editor.getAttributes('link').href as string;
    editor.chain().focus().extendMarkRange('link').run();
    const { from, to } = editor.state.selection;
    return editor.chain().insertContent(linkContent(href, editor.state.doc.textBetween(from, to), size));
  }

  function linkSizeHandler(size: LinkSize) {
    return command(
      (editor) => setLinkSize(editor, size),
      (editor) => (size === 'hyperlink' ? editor.isActive('link') : editor.isActive('linkCard', { size })),
    );
  }

  function linkHref(editor: Editor) {
    return (editor.getAttributes('linkCard').href ?? editor.getAttributes('link').href) as string | undefined;
  }

  // The notes offered after [[: picked, the link points to the note's id right away.
  const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
  const noteItems = computed(() =>
    (graph.value?.notes ?? [])
      .filter((note) => note.id !== id.value && note.title)
      .map((note) => ({ kind: 'wikiLink', label: note.title, description: note.path, icon: icons.note, id: note.id })),
  );

  function openLink(event: MouseEvent) {
    const link = (event.target as HTMLElement).closest<HTMLElement>('a[data-wikilink]');
    if (!link) return;
    event.preventDefault();
    if (link.dataset.note) return navigateTo(`/notes/${link.dataset.note}`);
    toast.add({
      title: t('note.pendingLink'),
      description: t('note.pendingLinkDescription', { target: link.dataset.wikilink }),
      color: 'warning',
    });
  }

  return {
    extensions: [WikiLink, linkCard(LinkCard)],
    starterKit: {
      link: {
        openOnClick: false,
        // Autolink asks with the text as typed: an email or a domain has no scheme yet, it gets mailto: or https:.
        isAllowedUri: (url: string, ctx: { defaultValidate: (url: string) => boolean }) => {
          const scheme = /^([a-z][a-z0-9+.-]*):/i.exec(url)?.[1];
          return (!scheme || ALLOWED_LINKS.includes(scheme.toLowerCase())) && ctx.defaultValidate(url);
        },
      },
    },
    handlers: {
      wikiLink: command((editor, item: { label: string; id: string }) =>
        editor
          .chain()
          .focus()
          .insertContent([
            { type: 'wikiLink', attrs: { target: item.label, id: item.id } },
            { type: 'text', text: ' ' },
          ]),
      ),
      link: command(
        (editor) => {
          askLink(editor);
          return editor.chain();
        },
        (editor) => editor.isActive('link'),
      ),
      linkHyperlink: linkSizeHandler('hyperlink'),
      linkMini: linkSizeHandler('mini'),
      linkLarge: linkSizeHandler('large'),
      linkOpen: command((editor) => {
        const href = linkHref(editor);
        // A mail link opens the mail app: a new tab would stay empty.
        if (href && isMailLink(href)) window.location.href = href;
        else if (isCardLink(href)) window.open(href, '_blank', 'noopener');
        return editor.chain();
      }),
    },
    // Shown on a web link or a card, like the menu of Apple Notes.
    bubbles: [
      {
        pluginKey: 'linkMenu',
        items: [
          [
            { kind: 'linkHyperlink', icon: icons.link, tooltip: { text: t('note.linkHyperlink') } },
            { kind: 'linkMini', icon: icons.linkMini, tooltip: { text: t('note.linkMini') } },
            { kind: 'linkLarge', icon: icons.linkLarge, tooltip: { text: t('note.linkLarge') } },
          ],
          [{ kind: 'linkOpen', icon: icons.linkOpen, tooltip: { text: t('note.linkOpen') } }],
        ],
        shouldShow: ({ editor }) =>
          editor.isActive('linkCard') || (editor.isActive('link') && isCardLink(editor.getAttributes('link').href)),
      },
    ],
    menus: [
      {
        char: '[[',
        pluginKey: 'wikiLinkMenu',
        items: noteItems,
        filterFields: ['label', 'description'],
        suggestion: { allowSpaces: true },
      },
    ],
    onClick: openLink,
  };
}
