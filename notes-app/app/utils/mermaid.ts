import { Extension } from '@tiptap/core';
import { Plugin, PluginKey, TextSelection, type EditorState } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';
import { watch } from 'vue';

const key = new PluginKey('mermaid');
let renders = 0;

/** The svg of a diagram, or its error; mermaid (~1 MB) is loaded with the first diagram. */
async function draw(code: string, dark: boolean) {
  const { default: mermaid } = await import('mermaid');
  // The svg comes from the text of the note: strict keeps its scripts and clicks out.
  mermaid.initialize({ startOnLoad: false, securityLevel: 'strict', theme: dark ? 'dark' : 'default' });
  return (await mermaid.render(`mermaid-${renders++}`, code)).svg;
}

function decorate(state: EditorState, dark: boolean, editable: boolean) {
  const { from, to } = state.selection;
  const decorations: Decoration[] = [];
  state.doc.descendants((node, pos) => {
    if (node.type.name !== 'codeBlock' || node.attrs.language !== 'mermaid') return;
    const end = pos + node.nodeSize;
    // The code is shown while the cursor is in it, the diagram below it; read only, never.
    if (!editable || from < pos || to > end) decorations.push(Decoration.node(pos, end, { class: 'mermaid-hidden' }));
    const code = node.textContent;
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
            (error: Error) => {
              box.classList.add('mermaid-error');
              box.textContent = error.message;
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

/** The ```mermaid code blocks drawn as diagrams, in the dark theme when dark says so. */
export const Mermaid = Extension.create<{ dark: () => boolean }>({
  name: 'mermaid',
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
