/** The API gives "" or "untitled" for a note with no title. */
export function displayTitle(title?: string) {
  return !title || title === 'untitled' ? 'Sans titre' : title;
}
