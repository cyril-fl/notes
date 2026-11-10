/** The LaTeX formulas of the note editor: $x^2$ edited by a click, $$…$$ a ```latex code block (useNoteCode draws it). */
export function useNoteMath(): EditorModule {
  const { prompt } = useDialogs();
  const { t } = useI18n();
  const edit = (latex: string) => prompt({ title: t('note.formula'), value: latex });
  // For everyone, or editing a note without them would read the _ and * of a formula as markdown.
  return { extensions: [InlineMath.configure({ edit }), BlockMath] };
}
