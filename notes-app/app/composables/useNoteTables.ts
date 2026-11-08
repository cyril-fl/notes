import { TableKit } from '@tiptap/extension-table';

/** The tables of the note editor: one inserted from the / menu, rows and columns added or removed from the bar over it. */
export function useNoteTables(): EditorModule[] {
  const { icons } = useIcons();
  const { t } = useI18n();

  return [
    // For everyone, or editing a note without them would escape the tables written elsewhere. Markdown has no column widths: nothing to resize.
    { extensions: [TableKit.configure({ table: { resizable: false } })] },
    {
      feature: 'tables',
      handlers: {
        // A header row: a markdown table needs one.
        table: command((editor) => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true })),
        tableAddRow: command((editor) => editor.chain().focus().addRowAfter()),
        tableAddColumn: command((editor) => editor.chain().focus().addColumnAfter()),
        tableDeleteRow: command((editor) => editor.chain().focus().deleteRow()),
        tableDeleteColumn: command((editor) => editor.chain().focus().deleteColumn()),
        tableDelete: command((editor) => editor.chain().focus().deleteTable()),
      },
      bubbles: [
        {
          pluginKey: 'tableMenu',
          items: [
            [
              { kind: 'tableAddRow', icon: icons.tableAddRow, tooltip: { text: t('note.tableAddRow') } },
              { kind: 'tableAddColumn', icon: icons.tableAddColumn, tooltip: { text: t('note.tableAddColumn') } },
            ],
            [
              { kind: 'tableDeleteRow', icon: icons.tableDeleteRow, tooltip: { text: t('note.tableDeleteRow') } },
              {
                kind: 'tableDeleteColumn',
                icon: icons.tableDeleteColumn,
                tooltip: { text: t('note.tableDeleteColumn') },
              },
              { kind: 'tableDelete', icon: icons.delete, tooltip: { text: t('note.tableDelete') } },
            ],
          ],
          shouldShow: ({ editor }) => editor.isActive('table'),
        },
      ],
      // Joins the / menu of useNoteSlash, or makes one if the account has only the tables.
      menus: [
        { char: '/', pluginKey: 'slashMenu', items: [{ kind: 'table', label: t('note.table'), icon: icons.table }] },
      ],
    },
  ];
}
