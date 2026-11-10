import { Extension, Node, mergeAttributes, nodeInputRule, textblockTypeInputRule } from '@tiptap/core';

// Pandoc's rule: no space inside the dollars, no digit after, so "$5 et $10" stays text.
const INLINE = /^\$(?=[^$\s])([^$\n]*?[^$\s])\$(?!\d)/;
// Empty too: an empty formula block is written $$ then $$.
const BLOCK = /^\$\$\n?([^$]*?)\n?\$\$[ \t]*(?:\n|$)/;

/** The $formula$ at the start of src, as a markdown token. */
export function readInlineMath(src: string) {
  const match = src.match(INLINE);
  return match ? { type: 'inlineMath', raw: match[0], latex: match[1]! } : undefined;
}

/** The $$formula$$ block at the start of src, on one line or several. */
export function readBlockMath(src: string) {
  const match = src.match(BLOCK);
  return match ? { type: 'blockMath', raw: match[0], latex: match[1]! } : undefined;
}

/** A formula drawn by KaTeX (~300 KB and its CSS), loaded with the first one; its source shown when it does not read. */
function draw(dom: HTMLElement, latex: string) {
  Promise.all([import('katex'), import('katex/dist/katex.min.css')]).then(([{ default: katex }]) => {
    try {
      katex.render(latex, dom);
    } catch {
      dom.textContent = latex;
      dom.classList.add('math-error');
    }
  });
}

type MathOptions = { edit: (latex: string) => Promise<string | undefined> };

/** The $formula$ of a line, kept as written in the markdown, edited by a click. */
export const InlineMath = Node.create<MathOptions>({
  name: 'inlineMath',
  inline: true,
  group: 'inline',
  atom: true,
  addOptions: () => ({ edit: async () => undefined }),
  addAttributes: () => ({ latex: { default: '', parseHTML: (el) => el.getAttribute('data-latex') } }),
  parseHTML: () => [{ tag: 'span[data-math]' }],
  renderHTML: ({ node, HTMLAttributes }) => [
    'span',
    mergeAttributes(HTMLAttributes, { 'data-math': '', 'data-latex': node.attrs.latex }),
  ],
  renderText: ({ node }) => `$${node.attrs.latex}$`,

  markdownTokenizer: {
    name: 'inlineMath',
    level: 'inline',
    start: (src) => src.indexOf('$'),
    tokenize: (src) => readInlineMath(src),
  },
  parseMarkdown: (token, helpers) => helpers.createNode('inlineMath', { latex: token.latex }),
  renderMarkdown: (node) => `$${node.attrs?.latex ?? ''}$`,

  addInputRules() {
    return [
      // No group in find: nodeInputRule would put the node in place of the group alone, dollars left around it.
      nodeInputRule({
        find: /(?<!\$)\$(?=[^$\s])[^$\n]*?[^$\s]\$$/,
        type: this.type,
        getAttributes: (match) => ({ latex: match[0].slice(1, -1).trim() }),
      }),
    ];
  },

  addNodeView() {
    return ({ node, editor, getPos }) => {
      const dom = document.createElement('span');
      dom.dataset.math = '';
      draw(dom, node.attrs.latex);
      // A click edits the source; the formula is not text the cursor goes in.
      dom.addEventListener('click', async () => {
        const pos = getPos();
        if (!editor.isEditable || pos === undefined) return;
        const latex = await this.options.edit(node.attrs.latex);
        if (latex)
          editor
            .chain()
            .focus()
            .command(({ tr }) => !!tr.setNodeMarkup(pos, undefined, { latex }))
            .run();
      });
      return { dom };
    };
  },
});

/**
 * The $$formula$$ blocks, a ```latex code block in the editor: written and drawn like a diagram (drawn.ts), $$ in
 * the markdown as in Obsidian. "$$ " at the start of a line makes one, as ``` does.
 */
export const BlockMath = Extension.create({
  name: 'blockMath',
  markdownTokenizer: {
    name: 'blockMath',
    level: 'block',
    // A block only at the start of a line: marked cuts a paragraph where start says.
    start: (src) => src.search(/^\$\$/m),
    tokenize: (src) => readBlockMath(src),
  },
  parseMarkdown: (token, helpers) =>
    helpers.createNode('codeBlock', { language: 'latex' }, token.latex ? [helpers.createTextNode(token.latex)] : []),
  addInputRules() {
    const type = this.editor.schema.nodes.codeBlock;
    return type ? [textblockTypeInputRule({ find: /^\$\$\s$/, type, getAttributes: { language: 'latex' } })] : [];
  },
});

/** The code block written back as $$ when it holds a formula. */
export function withBlockMath<T extends Node>(codeBlock: T) {
  return codeBlock.extend({
    renderMarkdown(node, helpers, context) {
      if (node.attrs?.language !== 'latex') return this.parent!(node, helpers, context);
      return `$$\n${helpers.renderChildren(node.content ?? [])}\n$$`;
    },
  });
}
