import { ALIGNMENTS, type EditorToolbarItem, MARKS } from '~/types/ui';

export const useEditorToolBar = (): EditorToolbarItem[][] => {
  const actions: EditorToolbarItem[] = [
    {
      kind: 'undo',
      icon: 'i-lucide-undo',
      tooltip: { label: 'editor.toolbar.undo' },
    },
    {
      kind: 'redo',
      icon: 'i-lucide-redo',
      tooltip: { label: 'editor.toolbar.redo' },
    },
  ];

  const formatting: EditorToolbarItem[] = [
    {
      kind: 'mark',
      mark: MARKS.BOLD,
      icon: 'i-lucide-bold',
      tooltip: { label: 'editor.toolbar.bold' },
    },
    {
      kind: 'mark',
      mark: MARKS.ITALIC,
      icon: 'i-lucide-italic',
      tooltip: { label: 'editor.toolbar.italic' },
    },
    {
      kind: 'mark',
      mark: MARKS.UNDERLINE,
      icon: 'i-lucide-underline',
      tooltip: { label: 'editor.toolbar.underline' },
    },
    {
      kind: 'mark',
      mark: MARKS.STRIKE,
      icon: 'i-lucide-strikethrough',
      tooltip: { label: 'editor.toolbar.strike' },
    },
  ];

  const headings: EditorToolbarItem[] = [
    {
      kind: 'heading',
      level: 1,
      icon: 'i-lucide-heading-1',
      tooltip: { label: 'editor.toolbar.heading1' },
    },
    {
      kind: 'heading',
      level: 2,
      icon: 'i-lucide-heading-2',
      tooltip: { label: 'editor.toolbar.heading2' },
    },
    {
      kind: 'heading',
      level: 3,
      icon: 'i-lucide-heading-3',
      tooltip: { label: 'editor.toolbar.heading3' },
    },
  ];

  const textAlignment: EditorToolbarItem[] = [
    {
      kind: 'textAlign',
      align: ALIGNMENTS.LEFT,
      icon: 'i-lucide-align-left',
      tooltip: { label: 'editor.toolbar.alignLeft' },
    },
    {
      kind: 'textAlign',
      align: ALIGNMENTS.CENTER,
      icon: 'i-lucide-align-center',
      tooltip: { label: 'editor.toolbar.alignCenter' },
    },
    {
      kind: 'textAlign',
      align: ALIGNMENTS.RIGHT,
      icon: 'i-lucide-align-right',
      tooltip: { label: 'editor.toolbar.alignRight' },
    },
  ];

  const lists: EditorToolbarItem[] = [
    {
      kind: 'bulletList',
      icon: 'i-lucide-list',
      tooltip: { label: 'editor.toolbar.bulletList' },
    },
    {
      kind: 'orderedList',
      icon: 'i-lucide-list-ordered',
      tooltip: { label: 'editor.toolbar.orderedList' },
    },
    {
      kind: 'blockquote',
      icon: 'i-lucide-text-quote',
      tooltip: { label: 'editor.toolbar.blockquote' },
    },
    {
      kind: 'mark',
      mark: MARKS.CODE,
      icon: 'i-lucide-code',
      tooltip: { label: 'editor.toolbar.code' },
    },
  ];

  const links: EditorToolbarItem[] = [
    {
      kind: 'link',
      icon: 'i-lucide-link',
      tooltip: { label: 'editor.toolbar.link' },
    },
  ];

  return [actions, formatting, headings, textAlignment, lists, links];
};
