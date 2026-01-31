import removeMarkdown from 'remove-markdown';

import { sha256 } from 'js-sha256';
import { useI18n } from 'vue-i18n';

/* Title extraction */
function truncateTitle(text: string): string {
  const config = useRuntimeConfig().public;

  const maxLength = config.notes.title.maxLength ?? 100;
  const toTruncate = text.length > maxLength;
  return toTruncate ? text.slice(0, maxLength) + '…' : text;
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
  const config = useRuntimeConfig().public;
  const { t } = useI18n();

  const title = findMarkdownTitle(markdownContent);
  return title ?? t(config.notes.title.default);
}

/* ID */
export function generateId(type: string): string {
  const raw = `${type}-${Date.now()}-${Math.floor(Math.random() * 10000)}`;
  return sha256(raw).substring(0, 12);
}
