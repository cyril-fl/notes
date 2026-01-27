import removeMarkdown from 'remove-markdown';

import { sha256 } from 'js-sha256';

/* Title extraction */
// TODO Mettre ca en param
const TITLE_MAX_LENGTH = 50;
const DEFAULT_TITLE = 'Sans titre';

function truncateTitle(text: string): string {
  const toTruncate = text.length > TITLE_MAX_LENGTH;
  return toTruncate ? text.slice(0, TITLE_MAX_LENGTH) + '…' : text;
}

function findMarkdownTitle(
  markdownContent: string,
  level: number = 1
): string | null {
  if (level > 6) return null;

  const hashes = '#'.repeat(level);
  const pattern = new RegExp(`^${hashes}\\s+(.+)$`, 'm');
  const match = markdownContent.match(pattern);

  if (match?.[1]) {
    const titleText = removeMarkdown(match[1]).trim();
    if (titleText) return truncateTitle(titleText);
  }

  return findMarkdownTitle(markdownContent, level + 1);
}

export function extractTitleFromContent(markdownContent: string): string {
  const title = findMarkdownTitle(markdownContent);
  return title ?? DEFAULT_TITLE;
}

/* ID */
export function generateId(type: string): string {
  const raw = `${type}-${Date.now()}-${Math.floor(Math.random() * 10000)}`;
  return sha256(raw).substring(0, 12);
}
