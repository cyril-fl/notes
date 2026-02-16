import { type EditorToolbarItem, MARKS } from '~/types/ui';
import LinkDialog from '~/components/UI/editor/LinkDialog.vue';

export const useEditorToolBar = (): EditorToolbarItem[][] => {
  const { editor } = useEditorContext();
  const overlay = useOverlay();
  const linkDialog = overlay.create(LinkDialog);
  const { icons } = useIcons();

  const actions: EditorToolbarItem[] = [
    {
      kind: 'undo',
      icon: icons.undo,
      tooltip: { label: 'editor.toolbar.undo' },
      onClick: () => editor.value?.chain().focus().undo().run(),
    },
    {
      kind: 'redo',
      icon: icons.redo,
      tooltip: { label: 'editor.toolbar.redo' },
      onClick: () => editor.value?.chain().focus().redo().run(),
    },
  ];

  const formatting: EditorToolbarItem[] = [
    {
      kind: 'mark',
      mark: MARKS.BOLD,
      icon: icons.bold,
      tooltip: { label: 'editor.toolbar.bold' },
      onClick: () => editor.value?.chain().focus().toggleBold().run(),
    },
    {
      kind: 'mark',
      mark: MARKS.ITALIC,
      icon: icons.italic,
      tooltip: { label: 'editor.toolbar.italic' },
      onClick: () => editor.value?.chain().focus().toggleItalic().run(),
    },
    {
      kind: 'mark',
      mark: MARKS.UNDERLINE,
      icon: icons.underline,
      tooltip: { label: 'editor.toolbar.underline' },
      onClick: () => editor.value?.chain().focus().toggleUnderline().run(),
    },
    {
      kind: 'mark',
      mark: MARKS.STRIKE,
      icon: icons.strike,
      tooltip: { label: 'editor.toolbar.strike' },
      onClick: () => editor.value?.chain().focus().toggleStrike().run(),
    },
    {
      kind: 'mark',
      mark: MARKS.HIGHLIGHT,
      icon: icons.highlight,
      tooltip: { label: 'editor.toolbar.highlight' },
      onClick: () => editor.value?.chain().focus().toggleHighlight().run(),
    },
  ];

  const headings: EditorToolbarItem[] = Array.from({ length: 6 }).map(
    (_, i) => {
      const level = (i + 1) as 1 | 2 | 3 | 4 | 5 | 6;
      return {
        kind: 'heading',
        level,
        icon: `i-lucide-heading-${level}`,
        tooltip: { label: `editor.toolbar.heading${level}` },
        onClick: () =>
          editor.value?.chain().focus().toggleHeading({ level }).run(),
      };
    }
  );

  const lists: EditorToolbarItem[] = [
    {
      kind: 'bulletList',
      icon: icons.unorderedlist,
      tooltip: { label: 'editor.toolbar.bulletList' },
      onClick: () => editor.value?.chain().focus().toggleBulletList().run(),
    },
    {
      kind: 'orderedList',
      icon: icons.orderlist,
      tooltip: { label: 'editor.toolbar.orderedList' },
      onClick: () => editor.value?.chain().focus().toggleOrderedList().run(),
    },
    {
      kind: 'taskList',
      icon: icons.tasklist,
      tooltip: { label: 'editor.toolbar.taskList' },
      onClick: () => editor.value?.chain().focus().toggleTaskList().run(),
    },
    {
      kind: 'blockquote',
      icon: icons.quote,
      tooltip: { label: 'editor.toolbar.blockquote' },
      onClick: () => editor.value?.chain().focus().toggleBlockquote().run(),
    },
  ];

  const blocks: EditorToolbarItem[] = [
    {
      kind: 'mark',
      mark: MARKS.CODE,
      icon: icons.code,
      tooltip: { label: 'editor.toolbar.code' },
      onClick: () => editor.value?.chain().focus().toggleCode().run(),
    },
    {
      kind: 'codeBlock',
      icon: icons.codeblock,
      tooltip: { label: 'editor.toolbar.codeBlock' },
      onClick: () => editor.value?.chain().focus().toggleCodeBlock().run(),
    },
    {
      kind: 'horizontalRule',
      icon: icons.horizontalRule,
      tooltip: { label: 'editor.toolbar.horizontalRule' },
      onClick: () => editor.value?.chain().focus().setHorizontalRule().run(),
    },
  ];

  const ALLOWED_LINK_PROTOCOLS: string[] = ['http', 'https', 'mailto'];

  const isAllowedLinkUrl = (url: string): boolean => {
    const trimmed = url.trim();
    if (!trimmed) return false;
    const protocol = (trimmed.split(':')[0] ?? '').toLowerCase();
    return ALLOWED_LINK_PROTOCOLS.includes(protocol);
  };

  const normalizeLinkHref = (input: string): string => {
    const trimmed = input.trim();
    if (!trimmed) return '';
    const protocol = (trimmed.split(':')[0] ?? '').toLowerCase();
    if (ALLOWED_LINK_PROTOCOLS.includes(protocol)) return trimmed;
    return `https://${trimmed}`;
  };

  const links: EditorToolbarItem[] = [
    {
      kind: 'link',
      icon: icons.link,
      tooltip: { label: 'editor.toolbar.link' },
      onClick: async () => {
        const ed = editor.value;
        if (!ed) return;
        const { href } = ed.getAttributes('link');
        const result = await linkDialog.open({ href: href ?? '' });
        if (!result) return;
        if ('unset' in result) {
          ed.chain().focus().unsetLink().run();
          return;
        }
        const linkHref = normalizeLinkHref(result.href);
        if (!isAllowedLinkUrl(linkHref)) return;
        ed.chain().focus().setLink({ href: linkHref }).run();
      },
    },
  ];

  return [actions, formatting, headings, lists, blocks, links];
};
