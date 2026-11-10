import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { common, createLowlight } from 'lowlight';

/** The languages offered over a code block; lowlight colors every other one of `common` written by hand (```rust). */
const LANGUAGES = {
  mermaid: 'Mermaid',
  latex: 'LaTeX',
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

/** The code blocks of the note editor colored by their language (```go), guessed when the block has none; ```mermaid and $$ (```latex) drawn. */
export function useNoteCode(): EditorModule {
  const colorMode = useColorMode();
  const { icons } = useIcons();
  const { t } = useI18n();
  return {
    feature: 'codeHighlight',
    // Replaces the starter kit's code block: same node and markdown, plus the colors.
    starterKit: { codeBlock: false },
    extensions: [
      withBlockMath(CodeBlockLowlight).configure({ lowlight: createLowlight(common) }),
      Drawn.configure({ dark: () => colorMode.value === 'dark' }),
    ],
    handlers: {
      codeLanguage: command(
        (editor, item) => editor.chain().focus().updateAttributes('codeBlock', { language: item.language }),
        (editor, item) => (editor.getAttributes('codeBlock').language ?? null) === item.language,
      ),
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
              ],
            },
          ],
        ],
        shouldShow: ({ editor }) => editor.isActive('codeBlock'),
      },
    ],
  };
}
