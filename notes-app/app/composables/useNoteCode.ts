import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { common, createLowlight } from 'lowlight';

/** The code blocks of the note editor colored by their language (```go), guessed when the block has none. */
export function useNoteCode(): EditorModule {
  return {
    feature: 'codeHighlight',
    // Replaces the starter kit's code block: same node and markdown, plus the colors.
    starterKit: { codeBlock: false },
    extensions: [CodeBlockLowlight.configure({ lowlight: createLowlight(common) })],
  };
}
