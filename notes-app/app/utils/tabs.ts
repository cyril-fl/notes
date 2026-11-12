/** The note to open when the tab of the open one is closed: the tab before it, else the one after. */
export function tabAfterClose(tabs: string[], closed: string): string | undefined {
  const index = tabs.indexOf(closed);
  if (index < 0) return undefined;
  return tabs[index - 1] ?? tabs[index + 1];
}
