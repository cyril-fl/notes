import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { common, createLowlight } from 'lowlight';

/** The code blocks of the note editor colored by their language (```go), guessed when the block has none; ```mermaid drawn. */
export function useNoteCode(): EditorModule {
  const colorMode = useColorMode();
  return {
    feature: 'codeHighlight',
    // Replaces the starter kit's code block: same node and markdown, plus the colors.
    starterKit: { codeBlock: false },
    extensions: [
      CodeBlockLowlight.configure({ lowlight: createLowlight(common) }),
      Mermaid.configure({ dark: () => colorMode.value === 'dark' }),
    ],
  };
}
