/** The menu after / in the note editor: turns the line into a heading, a list, a code block… with the handlers Nuxt UI already has, plus a formula and a diagram. */
export function useNoteSlash(): EditorModule {
  const { icons } = useIcons();
  const { t } = useI18n();
  const { prompt } = useDialogs();

  return {
    feature: 'slashCommands',
    handlers: {
      // The formula is asked first: an empty one would be an invisible node.
      formula: command((editor) => {
        prompt({ title: t('note.formula'), multiline: true }).then(
          (latex) => latex && editor.chain().focus().insertContent({ type: 'blockMath', attrs: { latex } }).run(),
        );
        return editor.chain();
      }),
      diagram: command((editor) => editor.chain().focus().setCodeBlock({ language: 'mermaid' })),
    },
    menus: [
      {
        char: '/',
        pluginKey: 'slashMenu',
        items: [
          { kind: 'paragraph', label: t('note.paragraph'), icon: icons.paragraph },
          {
            kind: 'heading',
            level: 1,
            group: t('note.headings'),
            label: t('note.heading', { level: 1 }),
            icon: icons.heading1,
          },
          {
            kind: 'heading',
            level: 2,
            group: t('note.headings'),
            label: t('note.heading', { level: 2 }),
            icon: icons.heading2,
          },
          {
            kind: 'heading',
            level: 3,
            group: t('note.headings'),
            label: t('note.heading', { level: 3 }),
            icon: icons.heading3,
          },
          { kind: 'bulletList', group: t('note.lists'), label: t('note.bulletList'), icon: icons.list },
          { kind: 'orderedList', group: t('note.lists'), label: t('note.orderedList'), icon: icons.orderedList },
          { kind: 'taskList', group: t('note.lists'), label: t('note.taskList'), icon: icons.taskList },
          { kind: 'blockquote', label: t('note.quote'), icon: icons.quote },
          { kind: 'codeBlock', label: t('note.codeBlock'), icon: icons.codeBlock },
          { kind: 'horizontalRule', label: t('note.divider'), icon: icons.divider },
          { kind: 'formula', label: t('note.formula'), icon: icons.formula },
          { kind: 'diagram', label: t('note.diagram'), icon: icons.diagram },
        ],
      },
    ],
  };
}
