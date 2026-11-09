/** The kinds of files the Storage settings sum up: a zip holds the Office documents, so it is a document. */
export const fileKinds = ['images', 'documents', 'audio', 'video'] as const;
export type FileKind = (typeof fileKinds)[number];

export const fileKind = (contentType: string): FileKind =>
  contentType.startsWith('image/')
    ? 'images'
    : contentType.startsWith('video/')
      ? 'video'
      : contentType.startsWith('audio/') || contentType === 'application/ogg'
        ? 'audio'
        : 'documents';

const units = ['byte', 'kilobyte', 'megabyte', 'gigabyte'];

/** A size in the units of the locale (Mo, MB), in powers of 1000 like Apple's storage. */
export function fileSize(bytes: number, locale: string) {
  let unit = 0;
  while (bytes >= 1000 && unit < units.length - 1) {
    bytes /= 1000;
    unit++;
  }
  return new Intl.NumberFormat(locale, {
    style: 'unit',
    unit: units[unit],
    unitDisplay: 'short',
    maximumFractionDigits: 1,
  }).format(bytes);
}
