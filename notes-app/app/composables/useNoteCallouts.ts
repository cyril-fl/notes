/** The callouts of the note editor (> [!note]): a colored box with its type, inserted from the / menu. */
export function useNoteCallouts(): EditorModule[] {
  const { icons } = useIcons();
  const { t } = useI18n();
  const types = {
    note: icons.calloutNote,
    tip: icons.calloutTip,
    warning: icons.calloutWarning,
    danger: icons.calloutDanger,
  };

  return [
    // For everyone, or editing a note without them would escape the callouts written elsewhere (> \[!note\]).
    // Replaces the starter kit's blockquote: a quote without a marker stays a quote.
    {
      starterKit: { blockquote: false },
      extensions: [
        Callout.configure({
          labels: Object.fromEntries(Object.keys(types).map((type) => [type, t(`note.callouts.${type}`)])),
        }),
      ],
    },
    {
      feature: 'callouts',
      handlers: {
        callout: command((editor, item) => editor.chain().focus().wrapIn('blockquote', { callout: item.callout })),
      },
      menus: [
        {
          char: '/',
          pluginKey: 'slashMenu',
          items: Object.entries(types).map(([callout, icon]) => ({
            kind: 'callout',
            callout,
            group: t('note.callout'),
            label: t(`note.callouts.${callout}`),
            icon,
          })),
        },
      ],
    },
  ];
}
