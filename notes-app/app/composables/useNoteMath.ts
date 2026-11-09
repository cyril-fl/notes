/** The LaTeX formulas of the note editor ($x^2$, $$…$$), edited by a click. */
export function useNoteMath(): EditorModule {
  const { prompt } = useDialogs();
  const { t } = useI18n();
  // TODO: the prompt is one line, a $$ block written on several lines comes back on one once edited; a textarea when one is asked for.
  const edit = (latex: string) => prompt({ title: t('note.formula'), value: latex });
  // For everyone, or editing a note without them would read the _ and * of a formula as markdown.
  return { extensions: [InlineMath.configure({ edit }), BlockMath.configure({ edit })] };
}
