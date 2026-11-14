/** The node of a tag: notes have uuids, no clash. The server names them the same (forcelayout.TagNode). */
export const tagNode = (tag: string) => `#${tag}`;

/**
 * The other named notes offered after [[, this one left out.
 * neighbors (Pro): the ones sharing the most tags with it first; the sort being stable, the API's order stays between equals.
 */
export function linkTargets<N extends { id: string; title: string }>(
  notes: N[] | null | undefined,
  id: string | undefined,
  neighbors?: { id: string; shared: number }[],
) {
  const shared = new Map(neighbors?.map((note) => [note.id, note.shared]));
  return (notes ?? [])
    .filter((note) => note.id !== id && note.title)
    .sort((a, b) => (shared.get(b.id) ?? 0) - (shared.get(a.id) ?? 0));
}
