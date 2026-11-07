/** The API gives "" or "untitled" for a note with no title: it shows as untitled, translated by the caller. */
export function displayTitle(title: string | undefined, untitled: string) {
  return !title || title === 'untitled' ? untitled : title;
}
