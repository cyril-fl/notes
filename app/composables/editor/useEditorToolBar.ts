import type { EditorToolbarItem } from '~/types/ui';

export const useEditorToolBar = (): EditorToolbarItem[][] => {
  const actions: EditorToolbarItem[] = [
    {
      kind: 'undo',
      icon: 'i-lucide-undo',
      tooltip: { text: 'Annuler' },
    },
    {
      kind: 'redo',
      icon: 'i-lucide-redo',
      tooltip: { text: 'Rétablir' },
    },
  ];

  const formatting: EditorToolbarItem[] = [
    {
      kind: 'mark',
      mark: 'bold',
      icon: 'i-lucide-bold',
      tooltip: { text: 'Gras' },
    },
    {
      kind: 'mark',
      mark: 'italic',
      icon: 'i-lucide-italic',
      tooltip: { text: 'Italique' },
    },
    {
      kind: 'mark',
      mark: 'underline',
      icon: 'i-lucide-underline',
      tooltip: { text: 'Souligné' },
    },
    {
      kind: 'mark',
      mark: 'strike',
      icon: 'i-lucide-strikethrough',
      tooltip: { text: 'Barré' },
    },
  ];

  const headings: EditorToolbarItem[] = [
    {
      kind: 'heading',
      level: 1,
      icon: 'i-lucide-heading-1',
      tooltip: { text: 'Titre 1' },
    },
    {
      kind: 'heading',
      level: 2,
      icon: 'i-lucide-heading-2',
      tooltip: { text: 'Titre 2' },
    },
    {
      kind: 'heading',
      level: 3,
      icon: 'i-lucide-heading-3',
      tooltip: { text: 'Titre 3' },
    },
  ];

  const textAlignment: EditorToolbarItem[] = [
    {
      kind: 'textAlign',
      align: 'left',
      icon: 'i-lucide-align-left',
      tooltip: { text: 'Aligner à gauche' },
    },
    {
      kind: 'textAlign',
      align: 'center',
      icon: 'i-lucide-align-center',
      tooltip: { text: 'Centrer' },
    },
    {
      kind: 'textAlign',
      align: 'right',
      icon: 'i-lucide-align-right',
      tooltip: { text: 'Aligner à droite' },
    },
  ];

  const lists: EditorToolbarItem[] = [
    {
      kind: 'bulletList',
      icon: 'i-lucide-list',
      tooltip: { text: 'Liste à puces' },
    },
    {
      kind: 'orderedList',
      icon: 'i-lucide-list-ordered',
      tooltip: { text: 'Liste numérotée' },
    },
    {
      kind: 'blockquote',
      icon: 'i-lucide-text-quote',
      tooltip: { text: 'Citation' },
    },
    {
      kind: 'mark',
      mark: 'code',
      icon: 'i-lucide-code',
      tooltip: { text: 'Code' },
    },
  ];

  const links: EditorToolbarItem[] = [
    {
      kind: 'link',
      icon: 'i-lucide-link',
      tooltip: { text: 'Lien' },
    },
  ];

  return [actions, formatting, headings, textAlignment, lists, links];
};
