import { Node, mergeAttributes, nodeInputRule } from '@tiptap/core';

// Pandoc's rule: no space inside the dollars, no digit after, so "$5 et $10" stays text.
const INLINE = /^\$(?=[^$\s])([^$\n]*?[^$\s])\$(?!\d)/;
const BLOCK = /^\$\$\n?([^$]+?)\n?\$\$[ \t]*(?:\n|$)/;

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
function draw(dom: HTMLElement, latex: string, displayMode: boolean) {
  Promise.all([import('katex'), import('katex/dist/katex.min.css')]).then(([{ default: katex }]) => {
    try {
      katex.render(latex, dom, { displayMode });
    } catch {
      dom.textContent = latex;
      dom.classList.add('math-error');
    }
  });
}

type MathOptions = { edit: (latex: string) => Promise<string | undefined> };

function mathNode(name: 'inlineMath' | 'blockMath', inline: boolean) {
  const [tag, write, read] = inline
    ? (['span', (latex: string) => `$${latex}$`, readInlineMath] as const)
    : (['div', (latex: string) => `$$\n${latex}\n$$`, readBlockMath] as const);
  return Node.create<MathOptions>({
    name,
    inline,
    group: inline ? 'inline' : 'block',
    atom: true,
    addOptions: () => ({ edit: async () => undefined }),
    addAttributes: () => ({ latex: { default: '', parseHTML: (el) => el.getAttribute('data-latex') } }),
    parseHTML: () => [{ tag: `${tag}[data-math]` }],
    renderHTML: ({ node, HTMLAttributes }) => [
      tag,
      mergeAttributes(HTMLAttributes, { 'data-math': '', 'data-latex': node.attrs.latex }),
    ],
    renderText: ({ node }) => write(node.attrs.latex),

    markdownTokenizer: {
      name,
      level: inline ? 'inline' : 'block',
      // A block only at the start of a line: marked cuts a paragraph where start says.
      start: (src) => (inline ? src.indexOf('$') : src.search(/^\$\$/m)),
      tokenize: (src) => read(src),
    },
    parseMarkdown: (token, helpers) => helpers.createNode(name, { latex: token.latex }),
    renderMarkdown: (node) => write(node.attrs?.latex ?? ''),

    addInputRules() {
      return [
        // No group in find: nodeInputRule would put the node in place of the group alone, dollars left around it.
        nodeInputRule({
          find: inline ? /(?<!\$)\$(?=[^$\s])[^$\n]*?[^$\s]\$$/ : /^\$\$[^$]+\$\$$/,
          type: this.type,
          getAttributes: (match) => ({ latex: match[0].slice(inline ? 1 : 2, inline ? -1 : -2).trim() }),
        }),
      ];
    },

    addNodeView() {
      return ({ node, editor, getPos }) => {
        const dom = document.createElement(tag);
        dom.dataset.math = '';
        draw(dom, node.attrs.latex, !inline);
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
}

/** The LaTeX formulas, $inline$ and $$block$$, kept as written in the markdown. */
export const InlineMath = mathNode('inlineMath', true);
export const BlockMath = mathNode('blockMath', false);
