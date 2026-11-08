import type { Component } from 'vue';
import { Plugin } from '@tiptap/pm/state';
import { mergeAttributes, Node, VueNodeViewRenderer } from '@tiptap/vue-3';

export type CardSize = 'mini' | 'large';
export type LinkSize = CardSize | 'hyperlink';

/** How a new web link shows, pasted or added with the link button. */
export const DEFAULT_LINK_SIZE: LinkSize = 'mini';

const CARD = /^\[([^[\]\n]*)\]\(((?:https?:\/\/|mailto:)[^\s)"]+) "(mini|large)"\)/i;

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

/** A link that can be a card: a web page, or an address to write to. */
export function isCardLink(href?: string | null) {
  return !!href && /^(https?:\/\/|mailto:)/i.test(href);
}

const EMAIL = /^[^\s@/:]+@[^\s@/:]+\.[^\s@/:]+$/;

/** The href of an address typed or pasted: an email gets mailto:, an address with no scheme https://. */
export function toHref(input: string) {
  const text = input.trim();
  if (EMAIL.test(text)) return `mailto:${text}`;
  return /^[a-z][a-z0-9+.-]*:/i.test(text) ? text : `https://${text}`;
}

export function isMailLink(href: string) {
  return /^mailto:/i.test(href);
}

/** The domain of a web link, the address of a mail link: its text when it has none. */
export function linkHost(href: string) {
  try {
    if (isMailLink(href)) return decodeURIComponent(href.slice(7).split('?')[0]!);
    return new URL(href).host;
  } catch {
    return href;
  }
}

/** The editor content of a link in a size: a card for a web or mail link, a text with the link mark otherwise. */
export function linkContent(href: string, text: string, size: LinkSize = DEFAULT_LINK_SIZE) {
  if (size === 'hyperlink' || !isCardLink(href))
    return { type: 'text', text: text || href, marks: [{ type: 'link', attrs: { href } }] };
  return { type: 'linkCard', attrs: { href, text: text || linkHost(href), size } };
}

/**
 * A web or mail link shown as a card, like in Apple Notes: mini sits in the line, large takes the whole width.
 * view draws it in the editor; copied, it is a plain <a>.
 */
export function linkCard(view: Component) {
  return Node.create({
    name: 'linkCard',
    group: 'inline',
    inline: true,
    atom: true,
    // Before the link mark's paste handler, which would make a hyperlink of a pasted address.
    priority: 1000,

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
            if (!isCardLink(href)) return false;
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

    // A pasted address or email becomes a link in the default size, on the selected text if any.
    addProseMirrorPlugins() {
      const { editor } = this;
      return [
        new Plugin({
          props: {
            handlePaste(view, event) {
              const text = event.clipboardData?.getData('text/plain').trim() ?? '';
              // Not a bare domain: example.com may be a word, the link mark's paste rule makes it a hyperlink.
              const href = EMAIL.test(text) ? toHref(text) : text;
              if (!isCardLink(href) || /\s/.test(href)) return false;
              const { from, to } = view.state.selection;
              editor
                .chain()
                .focus()
                .insertContent(linkContent(href, view.state.doc.textBetween(from, to)))
                .run();
              return true;
            },
          },
        }),
      ];
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
