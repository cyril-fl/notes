import { fitLayout, graphSignature, simulateGraph } from '../utils/graph';
import { readLayout, saveLayout } from '../utils/layoutCache';

type Graph = { notes: { id: string; tags?: string[] }[]; links: { source: string; target: string }[] };

// Off the page, which stays responsive: a big graph takes seconds to settle.
// cache: the key of the saved layout. The same graph comes back as saved; a changed one starts from it and only settles what changed.
onmessage = async ({ data: { notes, links, cache } }: MessageEvent<Graph & { cache?: string }>) => {
  const signature = graphSignature(notes, links);
  const saved = cache ? await readLayout(cache).catch(() => undefined) : undefined;
  const positions = saved?.signature === signature ? saved.positions : simulateGraph(notes, links, saved?.positions);
  postMessage(fitLayout(positions));
  if (cache && saved?.signature !== signature) await saveLayout(cache, { signature, positions }).catch(() => {});
  close();
};
