import { TaskItem, TaskList } from '@tiptap/extension-list';

/** The formatting of the note editor: headings, marks and lists, in the toolbar above the text. */
export function useNoteFormatting(): EditorModule {
  const { icons } = useIcons();
  const { t } = useI18n();

  return {
    extensions: [TaskList, TaskItem.configure({ nested: true })],
    starterKit: { underline: false },
    toolbar: [
      [
        { kind: 'heading', level: 1, icon: icons.heading1, tooltip: { text: t('note.heading', { level: 1 }) } },
        { kind: 'heading', level: 2, icon: icons.heading2, tooltip: { text: t('note.heading', { level: 2 }) } },
        { kind: 'heading', level: 3, icon: icons.heading3, tooltip: { text: t('note.heading', { level: 3 }) } },
      ],
      [
        { kind: 'mark', mark: 'bold', icon: icons.bold, tooltip: { text: t('note.bold') } },
        { kind: 'mark', mark: 'italic', icon: icons.italic, tooltip: { text: t('note.italic') } },
        { kind: 'mark', mark: 'strike', icon: icons.strike, tooltip: { text: t('note.strike') } },
        { kind: 'mark', mark: 'code', icon: icons.code, tooltip: { text: t('note.code') } },
        // Its handler comes from useNoteLinks.
        { kind: 'link', icon: icons.link, tooltip: { text: t('note.link') } },
      ],
      [
        { kind: 'bulletList', icon: icons.list, tooltip: { text: t('note.bulletList') } },
        { kind: 'orderedList', icon: icons.orderedList, tooltip: { text: t('note.orderedList') } },
        { kind: 'taskList', icon: icons.taskList, tooltip: { text: t('note.taskList') } },
        { kind: 'blockquote', icon: icons.quote, tooltip: { text: t('note.quote') } },
      ],
    ],
  };
}
