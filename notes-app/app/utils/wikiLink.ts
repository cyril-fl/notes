import { mergeAttributes, Node, nodeInputRule, nodePasteRule } from '@tiptap/vue-3';

const LINK = /\[\[([^[\]\n]+)\]\]/;

/** The [[target]] at the start of src, as the markdown lexer asks for it. */
export function readWikiLink(src: string) {
  const match = new RegExp(`^${LINK.source}`).exec(src);
  const target = match?.[1]?.trim();
  return match && target ? { type: 'wikiLink', raw: match[0], target } : undefined;
}

/**
 * A link to another note, written [[Title]] (or [[/Folder/Title]]) in the markdown: a node and not text,
 * or the markdown serializer would escape its brackets. Typing or pasting [[Title]] makes one.
 */
export const WikiLink = Node.create({
  name: 'wikiLink',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return { target: { default: '' } };
  },

  parseHTML() {
    return [{ tag: 'a[data-wikilink]', getAttrs: (el) => ({ target: (el as HTMLElement).dataset.wikilink }) }];
  },

  renderHTML({ node, HTMLAttributes }) {
    return ['a', mergeAttributes(HTMLAttributes, { 'data-wikilink': node.attrs.target, class: 'wikilink' }), node.attrs.target];
  },

  renderText({ node }) {
    return `[[${node.attrs.target}]]`;
  },

  markdownTokenName: 'wikiLink',
  markdownTokenizer: {
    name: 'wikiLink',
    level: 'inline',
    start: (src) => src.indexOf('[['),
    tokenize: (src) => readWikiLink(src),
  },
  parseMarkdown: (token, helpers) => helpers.createNode('wikiLink', { target: token.target }),
  renderMarkdown: (node) => `[[${node.attrs?.target}]]`,

  addInputRules() {
    return [nodeInputRule({ find: new RegExp(`${LINK.source}$`), type: this.type, getAttributes: (m) => ({ target: m[1]?.trim() }) })];
  },

  addPasteRules() {
    return [nodePasteRule({ find: new RegExp(LINK.source, 'g'), type: this.type, getAttributes: (m) => ({ target: m[1]?.trim() }) })];
  },
});
