const ENTITIES: Record<string, string> = { '&lt;': '<', '&gt;': '>', '&amp;': '&' };
// Code is split out (odd indexes): TipTap writes it as typed, an &lt; there is the text itself.
const CODE = /(^[ \t]*(?:```|~~~)[\s\S]*?(?:^[ \t]*(?:```|~~~)[^\n]*$|(?![\s\S]))|`[^`\n]*`)/m;

/**
 * Undoes the HTML entities TipTap's markdown serializer writes for & < > in text: the note stays plain markdown
 * for the CLI, the export and the search. One pass, so a typed "&lt;" (written "&amp;lt;") comes back as typed.
 */
export function decodeEntities(markdown: string) {
  return markdown
    .split(new RegExp(CODE.source, 'gm'))
    .map((part, index) => (index % 2 ? part : part.replace(/&(?:lt|gt|amp);/g, (entity) => ENTITIES[entity]!)))
    .join('');
}

/** The markdown of the editor as the note keeps it: it writes blank lines as non-breaking spaces, and & < > as entities.
 * Spaces first: a typed "&nbsp;" is written "&amp;nbsp;" and must come back as typed. */
export function cleanMarkdown(markdown: string) {
  return decodeEntities(markdown.replace(/&nbsp;/gi, ' ').replace(/\u00a0/g, ' ')).trim();
}
