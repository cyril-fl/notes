import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { common, createLowlight } from 'lowlight';

/** The languages offered over a code block; lowlight colors every other one of `common` written by hand (```rust). */
const LANGUAGES = {
  mermaid: 'Mermaid',
  go: 'Go',
  typescript: 'TypeScript',
  javascript: 'JavaScript',
  python: 'Python',
  sql: 'SQL',
  bash: 'Bash',
  json: 'JSON',
  yaml: 'YAML',
  html: 'HTML',
  css: 'CSS',
  markdown: 'Markdown',
};

/** The code blocks of the note editor colored by their language (```go), guessed when the block has none; ```mermaid drawn. */
export function useNoteCode(): EditorModule {
  const colorMode = useColorMode();
  const { icons } = useIcons();
  const { t } = useI18n();
  const { prompt } = useDialogs();
  return {
    feature: 'codeHighlight',
    // Replaces the starter kit's code block: same node and markdown, plus the colors.
    starterKit: { codeBlock: false },
    extensions: [
      CodeBlockLowlight.configure({ lowlight: createLowlight(common) }),
      Mermaid.configure({ dark: () => colorMode.value === 'dark' }),
    ],
    handlers: {
      codeLanguage: command(
        (editor, item) => editor.chain().focus().updateAttributes('codeBlock', { language: item.language }),
        (editor, item) => (editor.getAttributes('codeBlock').language ?? null) === item.language,
      ),
      // A formula is $$…$$ in Obsidian too, not ```latex: the block becomes one, its code as the formula. Asked
      // first, the code filled in: an empty one would be an invisible node.
      codeToFormula: command((editor) => {
        const pos = editor.state.selection.$from.before();
        prompt({ title: t('note.formula'), value: editor.state.doc.nodeAt(pos)?.textContent, multiline: true }).then(
          (latex) =>
            latex &&
            editor
              .chain()
              .focus()
              .command(({ state, tr }) => {
                const block = state.doc.nodeAt(pos);
                if (block?.type.name !== 'codeBlock') return false;
                tr.replaceWith(pos, pos + block.nodeSize, state.schema.nodes.blockMath!.create({ latex }));
                return true;
              })
              .run(),
        );
        return editor.chain();
      }),
    },
    bubbles: [
      {
        pluginKey: 'codeMenu',
        items: [
          [
            {
              icon: icons.codeBlock,
              label: t('note.codeLanguage'),
              tooltip: { text: t('note.codeLanguage') },
              items: [
                [{ kind: 'codeLanguage', language: null, label: t('note.codeAuto') }],
                Object.entries(LANGUAGES).map(([language, label]) => ({ kind: 'codeLanguage', language, label })),
                [{ kind: 'codeToFormula', label: t('note.formula'), icon: icons.formula }],
              ],
            },
          ],
        ],
        shouldShow: ({ editor }) => editor.isActive('codeBlock'),
      },
    ],
  };
}
