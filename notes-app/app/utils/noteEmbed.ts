import type { Component } from 'vue';
import { mergeAttributes, Node, nodeInputRule, nodePasteRule, VueNodeViewRenderer } from '@tiptap/vue-3';
import { readWikiLink, writeWikiLink } from './wikiLink';

const EMBED = /!\[\[([^[\]\n]+)\]\]/;

/** A note shown in another at the start of src, as the markdown lexer asks for it: ![[Title]] waiting, or ![Title](note:<id>). */
export function readNoteEmbed(src: string) {
  if (!src.startsWith('!')) return undefined;
  const link = readWikiLink(src.slice(1));
  return link && { ...link, type: 'noteEmbed', raw: `!${link.raw}` };
}

/** The markdown of an embed: a link with a ! before it, as Obsidian writes it; the export gives back ![[Title]]. */
export function writeNoteEmbed(target: string, id: string | null, heading: string | null = null) {
  return `!${writeWikiLink(target, id, heading)}`;
}

/** The text of an embedded note without the title heading it starts with: the link above it already gives it. */
export function withoutTitle(markdown: string) {
  return markdown.replace(/^\s*#{1,6}[ \t]+\S.*(?:\n|$)/, '').trimStart();
}

/**
 * A note shown in another, read only: the API points it to the note's id like a [[link]].
 * view draws the note; without it (inside an embed) it is a link, so A in B in A stops at one level.
 */
export function noteEmbed(view?: Component) {
  return Node.create({
    name: 'noteEmbed',
    group: 'inline',
    inline: true,
    atom: true,
    // Before the [[link]] rules, which would take the [[Title]] after the !.
    priority: 1000,

    addAttributes() {
      return { target: { default: '' }, id: { default: null }, heading: { default: null } };
    },

    parseHTML() {
      return [
        {
          tag: 'a[data-embed]',
          getAttrs: (el) => ({
            target: (el as HTMLElement).dataset.embed,
            id: (el as HTMLElement).dataset.note ?? null,
            heading: (el as HTMLElement).dataset.heading ?? null,
          }),
        },
      ];
    },

    // A link the click handler of the editor opens, like a [[link]].
    renderHTML({ node, HTMLAttributes }) {
      const attrs = {
        'data-embed': node.attrs.target,
        'data-wikilink': node.attrs.target,
        'data-note': node.attrs.id ?? undefined,
        'data-heading': node.attrs.heading ?? undefined,
        class: node.attrs.id ? 'wikilink' : 'wikilink wikilink-waiting',
      };
      return ['a', mergeAttributes(HTMLAttributes, attrs), node.attrs.target];
    },

    renderText({ node }) {
      return writeNoteEmbed(node.attrs.target, node.attrs.id, node.attrs.heading);
    },

    ...(view && { addNodeView: () => VueNodeViewRenderer(view) }),

    markdownTokenName: 'noteEmbed',
    markdownTokenizer: {
      name: 'noteEmbed',
      level: 'inline',
      start: (src) => src.indexOf('!['),
      tokenize: (src) => readNoteEmbed(src),
    },
    parseMarkdown: (token, helpers) =>
      helpers.createNode('noteEmbed', { target: token.target, id: token.id, heading: token.heading }),
    renderMarkdown: (node) => writeNoteEmbed(node.attrs?.target, node.attrs?.id ?? null, node.attrs?.heading ?? null),

    addInputRules() {
      return [
        // No group in find: nodeInputRule would put the node in place of the group alone, brackets left around it.
        nodeInputRule({
          find: new RegExp(`${EMBED.source.replace('(', '(?:')}$`),
          type: this.type,
          getAttributes: (m) => ({ target: m[0].slice(3, -2).trim() }),
        }),
      ];
    },

    addPasteRules() {
      return [
        nodePasteRule({
          find: new RegExp(EMBED.source, 'g'),
          type: this.type,
          getAttributes: (m) => ({ target: m[1]?.trim() }),
        }),
      ];
    },
  });
}
