import { layoutGraph } from '../utils/graph';

// Off the page, which stays responsive: a big graph takes seconds to settle.
onmessage = ({
  data,
}: MessageEvent<{ notes: { id: string; tags?: string[] }[]; links: { source: string; target: string }[] }>) =>
  postMessage(layoutGraph(data.notes, data.links));
