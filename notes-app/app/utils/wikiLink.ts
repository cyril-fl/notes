import { mergeAttributes, Node, nodeInputRule, nodePasteRule } from '@tiptap/vue-3';

const BY_TITLE = /\[\[([^[\]\n]+)\]\]/;
const BY_ID = /\[([^[\]\n]*)\]\(note:([0-9a-fA-F-]{36})(?:#([^()\n]*))?\)/;

/**
 * The link to a note at the start of src, as the markdown lexer asks for it: [[Title]] waiting, or [Title](note:<id>);
 * [Title#Heading](note:<id>#Heading) points to a heading of the note.
 */
export function readWikiLink(src: string) {
  const byId = new RegExp(`^${BY_ID.source}`).exec(src);
  if (byId)
    return {
      type: 'wikiLink',
      raw: byId[0],
      target: byId[1]!.trim(),
      id: byId[2]!.toLowerCase(),
      heading: byId[3]?.trim() || null,
    };
  const byTitle = new RegExp(`^${BY_TITLE.source}`).exec(src);
  const target = byTitle?.[1]?.trim();
  return byTitle && target ? { type: 'wikiLink', raw: byTitle[0], target, id: null, heading: null } : undefined;
}

/** The markdown of a link: by id once the API resolved it, by title until then. */
export function writeWikiLink(target: string, id: string | null, heading: string | null = null) {
  if (!id) return `[[${target}]]`;
  return `[${target.replace(/[[\]]/g, '')}](note:${id}${heading ? `#${heading.replace(/[()\n]/g, '')}` : ''})`;
}

/** A heading as a link and the editor compare it: the link drops its parentheses, whatever the case. */
export function sameHeading(a: string, b: string) {
  const plain = (heading: string) => heading.replace(/[()]/g, '').trim().toLowerCase();
  return plain(a) === plain(b);
}

/** The headings of a markdown text, in order, those of code blocks left out. */
export function headingsOf(markdown: string) {
  let code = false;
  const headings: string[] = [];
  for (const line of markdown.split('\n')) {
    if (/^\s*```/.test(line)) code = !code;
    const heading = !code && /^#{1,6}\s+(.+?)\s*$/.exec(line)?.[1];
    if (heading) headings.push(heading);
  }
  return headings;
}

/** The ids of the notes a saved text links to, sorted: two texts link to the same notes when their lists are equal. */
export function linkedNotes(content: string) {
  const ids = [...content.matchAll(new RegExp(BY_ID.source, 'g'))].map((link) => link[2]!.toLowerCase());
  return [...new Set(ids)].sort().join();
}

/** The [[targets]] the API resolved on saving, by lowercased target: it rewrites each link in place, so sent and saved links pair up in order. */
export function resolvedLinks(sent: string, saved: string) {
  const links = (content: string) => [...content.matchAll(new RegExp(`${BY_ID.source}|${BY_TITLE.source}`, 'g'))];
  const before = links(sent);
  const after = links(saved);
  const resolved = new Map<string, { target: string; id: string; heading: string | null }>();
  if (before.length !== after.length) return resolved;
  before.forEach((link, i) => {
    const target = link[4]?.trim();
    const [, label, id, heading] = after[i]!;
    if (target && id)
      resolved.set(target.toLowerCase(), {
        target: label!.trim(),
        id: id.toLowerCase(),
        heading: heading?.trim() || null,
      });
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
    return { target: { default: '' }, id: { default: null }, heading: { default: null } };
  },

  parseHTML() {
    return [
      {
        tag: 'a[data-wikilink]',
        getAttrs: (el) => ({
          target: (el as HTMLElement).dataset.wikilink,
          id: (el as HTMLElement).dataset.note ?? null,
          heading: (el as HTMLElement).dataset.heading ?? null,
        }),
      },
    ];
  },

  renderHTML({ node, HTMLAttributes }) {
    const attrs = {
      'data-wikilink': node.attrs.target,
      'data-note': node.attrs.id ?? undefined,
      'data-heading': node.attrs.heading ?? undefined,
      // [[#Heading]] points to the note itself: it has no id to wait for.
      class: node.attrs.id || node.attrs.target.startsWith('#') ? 'wikilink' : 'wikilink wikilink-waiting',
    };
    return ['a', mergeAttributes(HTMLAttributes, attrs), node.attrs.target];
  },

  renderText({ node }) {
    return writeWikiLink(node.attrs.target, node.attrs.id, node.attrs.heading);
  },

  markdownTokenName: 'wikiLink',
  markdownTokenizer: {
    name: 'wikiLink',
    level: 'inline',
    start: (src) => src.indexOf('['),
    tokenize: (src) => readWikiLink(src),
  },
  parseMarkdown: (token, helpers) =>
    helpers.createNode('wikiLink', { target: token.target, id: token.id, heading: token.heading }),
  renderMarkdown: (node) => writeWikiLink(node.attrs?.target, node.attrs?.id ?? null, node.attrs?.heading ?? null),

  addInputRules() {
    return [
      // No group in find: nodeInputRule would put the node in place of the group alone, brackets left around it.
      nodeInputRule({
        find: new RegExp(`${BY_TITLE.source.replace('(', '(?:')}$`),
        type: this.type,
        getAttributes: (m) => ({ target: m[0].slice(2, -2).trim() }),
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
