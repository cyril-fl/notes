import { forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { SimulationNodeDatum } from 'd3-force';

type Node = SimulationNodeDatum & { id: string };
type Link = { source: string; target: string; tags: string[]; linked?: boolean };

const TICKS = 300;
const MARGIN = 8;
const COLLIDE_RADIUS = 30;
// In % of the view per unit of the simulation: keeps a small graph from spreading to the corners.
const MAX_SCALE = 0.35;

/** Scales the simulated values into the view, around its center. */
function fit(values: number[]) {
  const min = Math.min(...values);
  const range = Math.max(...values) - min;
  const scale = range ? Math.min((100 - 2 * MARGIN) / range, MAX_SCALE) : 0;
  return (value: number) => 50 + (value - min - range / 2) * scale;
}

/**
 * Where the notes go, in % of the view: the more tags two notes share, the closer they are; a [[link]] counts as one.
 * d3 seeds its randomness, so the same notes get the same map on every load. GraphView zooms by spreading them.
 */
export function layoutGraph(notes: { id: string }[], links: Link[]) {
  const nodes: Node[] = notes.map(({ id }) => ({ id }));
  // d3 writes into the links it gets: they are copies.
  const edges = links.map((link) => ({ ...link }));

  forceSimulation(nodes)
    .force('charge', forceManyBody().strength(-300))
    .force(
      'link',
      forceLink<Node, Link>(edges)
        .id((node) => node.id)
        .distance((link) => 70 / Math.max(link.tags.length + (link.linked ? 1 : 0), 1))
        .strength(1),
    )
    .force('collide', forceCollide(COLLIDE_RADIUS))
    .force('x', forceX().strength(0.04))
    .force('y', forceY().strength(0.08))
    .stop()
    .tick(TICKS);

  const x = fit(nodes.map((node) => node.x!));
  const y = fit(nodes.map((node) => node.y!));
  return new Map(nodes.map((node) => [node.id, { x: x(node.x!), y: y(node.y!) }]));
}

/** The graph around one note: the note, the notes it shares a [[link]] or a tag with, and the links between them. */
export function neighborhood<G extends { notes: { id: string }[]; links: Link[] }>(graph: G, id: string): G {
  const ids = new Set([id]);
  for (const link of graph.links) if (link.source === id || link.target === id) ids.add(link.source).add(link.target);
  return {
    ...graph,
    notes: graph.notes.filter((note) => ids.has(note.id)),
    links: graph.links.filter((link) => ids.has(link.source) && ids.has(link.target)),
  };
}
