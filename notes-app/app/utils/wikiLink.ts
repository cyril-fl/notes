import { mergeAttributes, Node, nodeInputRule, nodePasteRule } from '@tiptap/vue-3';

const BY_TITLE = /\[\[([^[\]\n]+)\]\]/;
const BY_ID = /\[([^[\]\n]*)\]\(note:([0-9a-fA-F-]{36})\)/;

/** The link to a note at the start of src, as the markdown lexer asks for it: [[Title]] waiting, or [Title](note:<id>). */
export function readWikiLink(src: string) {
  const byId = new RegExp(`^${BY_ID.source}`).exec(src);
  if (byId) return { type: 'wikiLink', raw: byId[0], target: byId[1]!.trim(), id: byId[2]!.toLowerCase() };
  const byTitle = new RegExp(`^${BY_TITLE.source}`).exec(src);
  const target = byTitle?.[1]?.trim();
  return byTitle && target ? { type: 'wikiLink', raw: byTitle[0], target, id: null } : undefined;
}

/** The markdown of a link: by id once the API resolved it, by title until then. */
export function writeWikiLink(target: string, id: string | null) {
  return id ? `[${target.replace(/[[\]]/g, '')}](note:${id})` : `[[${target}]]`;
}

/** The [[targets]] the API resolved on saving, by lowercased target: it rewrites each link in place, so sent and saved links pair up in order. */
export function resolvedLinks(sent: string, saved: string) {
  const links = (content: string) => [...content.matchAll(new RegExp(`${BY_ID.source}|${BY_TITLE.source}`, 'g'))];
  const before = links(sent);
  const after = links(saved);
  const resolved = new Map<string, { target: string; id: string }>();
  if (before.length !== after.length) return resolved;
  before.forEach((link, i) => {
    const target = link[3]?.trim();
    const id = after[i]![2];
    if (target && id) resolved.set(target.toLowerCase(), { target: after[i]![1]!.trim(), id: id.toLowerCase() });
  });
  return resolved;
}

/**
 * A link to another note: a node and not text, or the markdown serializer would escape its brackets.
 * Typing or pasting [[Title]] makes one, the API then points it to the note's id.
 */
export const WikiLink = Node.create({
  name: 'wikiLink',
  group: 'inline',
  inline: true,
  atom: true,

  addAttributes() {
    return { target: { default: '' }, id: { default: null } };
  },

  parseHTML() {
    return [
      {
        tag: 'a[data-wikilink]',
        getAttrs: (el) => ({
          target: (el as HTMLElement).dataset.wikilink,
          id: (el as HTMLElement).dataset.note ?? null,
        }),
      },
    ];
  },

  renderHTML({ node, HTMLAttributes }) {
    const attrs = {
      'data-wikilink': node.attrs.target,
      'data-note': node.attrs.id ?? undefined,
      class: node.attrs.id ? 'wikilink' : 'wikilink wikilink-waiting',
    };
    return ['a', mergeAttributes(HTMLAttributes, attrs), node.attrs.target];
  },

  renderText({ node }) {
    return writeWikiLink(node.attrs.target, node.attrs.id);
  },

  markdownTokenName: 'wikiLink',
  markdownTokenizer: {
    name: 'wikiLink',
    level: 'inline',
    start: (src) => src.indexOf('['),
    tokenize: (src) => readWikiLink(src),
  },
  parseMarkdown: (token, helpers) => helpers.createNode('wikiLink', { target: token.target, id: token.id }),
  renderMarkdown: (node) => writeWikiLink(node.attrs?.target, node.attrs?.id ?? null),

  addInputRules() {
    return [
      nodeInputRule({
        find: new RegExp(`${BY_TITLE.source}$`),
        type: this.type,
        getAttributes: (m) => ({ target: m[1]?.trim() }),
      }),
    ];
  },

  addPasteRules() {
    return [
      nodePasteRule({
        find: new RegExp(BY_TITLE.source, 'g'),
        type: this.type,
        getAttributes: (m) => ({ target: m[1]?.trim() }),
      }),
    ];
  },
});
