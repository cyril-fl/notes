import type { Component } from 'vue';
import { mergeAttributes, Node, VueNodeViewRenderer } from '@tiptap/vue-3';

export type CardSize = 'mini' | 'large';

const CARD = /^\[([^[\]\n]*)\]\((https?:\/\/[^\s)"]+) "(mini|large)"\)/i;

/** A web link shown as a card at the start of src, as the markdown lexer asks for it: [Text](https://… "mini" or "large"). */
export function readLinkCard(src: string) {
  const card = CARD.exec(src);
  if (!card) return undefined;
  return {
    type: 'linkCard',
    raw: card[0],
    text: card[1]!.trim(),
    href: card[2]!,
    size: card[3]!.toLowerCase() as CardSize,
  };
}

/** The markdown of a card: a standard link whose title gives the size, so it stays a plain link anywhere else. */
export function writeLinkCard(text: string, href: string, size: CardSize) {
  return `[${text.replace(/[[\]]/g, '')}](${href.replace(/[\s()"]/g, (c) => `%${c.charCodeAt(0).toString(16).toUpperCase().padStart(2, '0')}`)} "${size}")`;
}

export function isWebLink(href?: string | null) {
  return !!href && /^https?:\/\//i.test(href);
}

/**
 * A web link shown as a card, like in Apple Notes: mini sits in the line, large takes the whole width.
 * view draws it in the editor; copied, it is a plain <a>.
 */
export function linkCard(view: Component) {
  return Node.create({
    name: 'linkCard',
    group: 'inline',
    inline: true,
    atom: true,

    addAttributes() {
      return { href: { default: '' }, text: { default: '' }, size: { default: 'mini' } };
    },

    parseHTML() {
      return [
        {
          tag: 'a[data-link-card]',
          getAttrs: (el) => {
            const a = el as HTMLAnchorElement;
            const href = a.getAttribute('href');
            if (!isWebLink(href)) return false;
            return { href, text: a.textContent ?? '', size: a.dataset.linkCard === 'large' ? 'large' : 'mini' };
          },
        },
      ];
    },

    renderHTML({ node, HTMLAttributes }) {
      const attrs = { href: node.attrs.href, 'data-link-card': node.attrs.size };
      return ['a', mergeAttributes(HTMLAttributes, attrs), node.attrs.text || node.attrs.href];
    },

    renderText({ node }) {
      return writeLinkCard(node.attrs.text, node.attrs.href, node.attrs.size);
    },

    addNodeView() {
      return VueNodeViewRenderer(view);
    },

    markdownTokenName: 'linkCard',
    markdownTokenizer: {
      name: 'linkCard',
      level: 'inline',
      start: (src) => src.indexOf('['),
      tokenize: (src) => readLinkCard(src),
    },
    parseMarkdown: (token, helpers) =>
      helpers.createNode('linkCard', { href: token.href, text: token.text, size: token.size }),
    renderMarkdown: (node) => writeLinkCard(node.attrs?.text ?? '', node.attrs?.href ?? '', node.attrs?.size ?? 'mini'),
  });
}
