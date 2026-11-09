/** The LaTeX formulas of the note editor ($x^2$, $$…$$), edited by a click. */
export function useNoteMath(): EditorModule {
  const { prompt } = useDialogs();
  const { t } = useI18n();
  const edit = (multiline: boolean) => (latex: string) => prompt({ title: t('note.formula'), value: latex, multiline });
  // For everyone, or editing a note without them would read the _ and * of a formula as markdown.
  return { extensions: [InlineMath.configure({ edit: edit(false) }), BlockMath.configure({ edit: edit(true) })] };
}
