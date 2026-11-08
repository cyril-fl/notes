import { Extension, type Range } from '@tiptap/core';
import { PluginKey } from '@tiptap/pm/state';
import Suggestion from '@tiptap/suggestion';

/** The / just typed at the start of a word: open gets the range to delete and where to show a menu. */
export const slashTrigger = (open: (range: Range, rect: DOMRect | null) => void) =>
  Extension.create({
    name: 'slashTrigger',
    addProseMirrorPlugins() {
      return [
        Suggestion({
          editor: this.editor,
          char: '/',
          pluginKey: new PluginKey('slashTrigger'),
          // Not in code, where a / is text: // would open it.
          allow: ({ state, range }) => !state.doc.resolve(range.from).parent.type.spec.code,
          // Only a / just typed: a click or an arrow after a / already in the text does not open it.
          shouldShow: ({ transaction }) => transaction.docChanged,
          // Only a bare /: the cursor coming back after a /word does not open it again.
          render: () => ({ onStart: (props) => props.query === '' && open(props.range, props.clientRect?.() ?? null) }),
        }),
      ];
    },
  });
