import { Blockquote } from '@tiptap/extension-blockquote';
import type { JSONContent } from '@tiptap/core';

const MARKER = /^\[!([a-z-]+)\][ \t]*/i;

/** The callouts of Obsidian and GitHub (> [!note]): a blockquote that keeps its type, and its title when the marker has one on its line. */
export const Callout = Blockquote.extend<{ labels: Record<string, string> }>({
  addOptions() {
    return { ...this.parent!(), labels: {} };
  },

  addAttributes() {
    return {
      callout: {
        default: null,
        parseHTML: (element) => element.getAttribute('data-callout'),
        // The type named in the account's language above the callout, as written when it has no label.
        renderHTML: (attributes) =>
          attributes.callout
            ? {
                'data-callout': attributes.callout,
                'data-label': this.options.labels[attributes.callout.toLowerCase()] ?? attributes.callout,
              }
            : {},
      },
      titled: { default: false, rendered: false },
    };
  },

  parseMarkdown(token, helpers) {
    const node = this.parent!(token, helpers) as JSONContent;
    const first = node.content?.[0];
    const text = first?.type === 'paragraph' ? first.content?.[0] : undefined;
    const marker = text?.type === 'text' && text.marks === undefined ? text.text?.match(MARKER) : null;
    if (!marker) return node;

    const rest = text!.text!.slice(marker[0].length);
    node.attrs = { ...node.attrs, callout: marker[1], titled: rest !== '' && !rest.startsWith('\n') };
    const after = rest.replace(/^\n/, '');
    if (after) text!.text = after;
    else first!.content!.shift();
    // A marker alone in its paragraph leaves nothing to show.
    if (!first!.content!.length && node.content!.length > 1) node.content!.shift();
    return node;
  },

  renderMarkdown(node, helpers, context) {
    const quote = this.parent!(node, helpers, context) as string;
    const { callout, titled } = node.attrs ?? {};
    if (!callout) return quote;
    return titled ? quote.replace(/^> ?/, `> [!${callout}] `) : `> [!${callout}]\n${quote}`;
  },
});
