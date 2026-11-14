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
  latex: 'LaTeX',
};

/** lowlight, whose guess colors a block only when sure: a word like `dd` was guessed bash, a sentence kotlin. */
function sureLowlight() {
  const lowlight = createLowlight(common);
  const guess = lowlight.highlightAuto;
  lowlight.highlightAuto = (value, options) => {
    const tree = guess(value, options);
    // Real code of a few lines scores 5 and more, a guess on a few words 1 to 3.
    return (tree.data?.relevance ?? 0) >= 5 ? tree : { type: 'root', children: [{ type: 'text', value }] };
  };
  return lowlight;
}

/** The code blocks of the note editor colored by their language (```go), guessed when the block has none; ```mermaid and $$ (```math) drawn. */
export function useNoteCode(): EditorModule {
  const colorMode = useColorMode();
  const { icons } = useIcons();
  const { t } = useI18n();
  return {
    feature: 'codeHighlight',
    // Replaces the starter kit's code block: same node and markdown, plus the colors.
    starterKit: { codeBlock: false },
    extensions: [
      withBlockMath(CodeBlockLowlight).configure({ lowlight: sureLowlight() }),
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
              // Closed, the menu gave the focus back to its button, after the editor's: the first key typed was lost.
              content: { onCloseAutoFocus: (event: Event) => event.preventDefault() },
              items: [
                [{ kind: 'codeLanguage', language: null, label: t('note.codeAuto') }],
                [
                  // The formula, `math` and not `latex`, which stays a listing of LaTeX source (#124).
                  { kind: 'codeLanguage', language: 'math', label: t('note.formula') },
                  ...Object.entries(LANGUAGES).map(([language, label]) => ({ kind: 'codeLanguage', language, label })),
                ],
              ],
            },
          ],
        ],
        shouldShow: ({ editor }) => editor.isActive('codeBlock'),
      },
    ],
  };
}
