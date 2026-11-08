/** The API gives "" or "untitled" for a note with no title: it shows as untitled, translated by the caller. */
export function displayTitle(title: string | undefined, untitled: string) {
  return !title || title === 'untitled' ? untitled : title;
}

const fence = /^[ \t]*(```|~~~)/;
const heading = /^ {0,3}#{1,6}[ \t]+(?=\S)/;

/** The markdown with title in its first heading, as the API reads it (code skipped), else a heading added on top. */
export function retitle(markdown: string, title: string) {
  const lines = markdown.split('\n');
  let code = false;
  for (const [index, line] of lines.entries()) {
    if (fence.test(line)) code = !code;
    const marks = !code && line.match(heading)?.[0];
    if (marks) {
      lines[index] = marks + title;
      return lines.join('\n');
    }
  }
  return markdown ? `# ${title}\n\n${markdown}` : `# ${title}`;
}
