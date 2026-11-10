import { Extension } from '@tiptap/core';
import { Plugin, PluginKey, TextSelection, type EditorState } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';
import { watch } from 'vue';

const key = new PluginKey('drawn');
let renders = 0;

/** The html of a block's code by its language, or its error; each library loaded with its first block. */
const drawers: Record<string, (code: string, dark: boolean) => Promise<string>> = {
  // mermaid is ~1 MB.
  async mermaid(code, dark) {
    const { default: mermaid } = await import('mermaid');
    // The svg comes from the text of the note: strict keeps its scripts and clicks out.
    mermaid.initialize({ startOnLoad: false, securityLevel: 'strict', theme: dark ? 'dark' : 'default' });
    return (await mermaid.render(`mermaid-${renders++}`, code)).svg;
  },
  // KaTeX is ~300 KB and its CSS.
  async latex(code) {
    const [{ default: katex }] = await Promise.all([import('katex'), import('katex/dist/katex.min.css')]);
    return katex.renderToString(code, { displayMode: true });
  },
};

function decorate(state: EditorState, dark: boolean, editable: boolean) {
  const { from, to } = state.selection;
  const decorations: Decoration[] = [];
  state.doc.descendants((node, pos) => {
    const draw = node.type.name === 'codeBlock' && drawers[node.attrs.language];
    const code = node.textContent;
    // An empty one stays a code block to write in: nothing to draw, it would vanish.
    if (!draw || !code.trim()) return;
    const end = pos + node.nodeSize;
    // The code is shown while the cursor is in it, the drawing below it; read only, never.
    if (!editable || from < pos || to > end) decorations.push(Decoration.node(pos, end, { class: 'mermaid-hidden' }));
    decorations.push(
      Decoration.widget(
        end,
        (view, getPos) => {
          const box = document.createElement('div');
          box.className = 'mermaid-diagram';
          box.contentEditable = 'false';
          // A click puts the cursor at the end of the code, which shows it.
          box.addEventListener('mousedown', (event) => {
            const at = getPos();
            if (at === undefined || !view.editable) return;
            event.preventDefault();
            view.dispatch(view.state.tr.setSelection(TextSelection.create(view.state.doc, at - 1)));
            view.focus();
          });
          draw(code, dark).then(
            (svg) => (box.innerHTML = svg),
            // In the block's box, a box of its own: the drawing failed, not the block.
            (error: Error) => {
              const message = document.createElement('div');
              message.className = 'mermaid-error';
              message.textContent = error.message;
              box.replaceChildren(message);
            },
          );
          return box;
        },
        // Kept as long as its code and the mode are the same: not drawn again on each key typed elsewhere.
        { side: -1, key: `${dark}:${code}` },
      ),
    );
  });
  return DecorationSet.create(state.doc, decorations);
}

/** The ```mermaid code blocks drawn as diagrams, in the dark theme when dark says so, the ```latex ones as formulas. */
export const Drawn = Extension.create<{ dark: () => boolean }>({
  name: 'drawn',
  addOptions: () => ({ dark: () => false }),
  addProseMirrorPlugins() {
    const { dark } = this.options;
    const editor = this.editor;
    return [
      new Plugin({
        key,
        props: { decorations: (state) => decorate(state, dark(), editor.isEditable) },
        // A change of mode draws them again.
        view: (view) => ({ destroy: watch(dark, () => view.dispatch(view.state.tr.setMeta(key, true))) }),
      }),
    ];
  },
});
