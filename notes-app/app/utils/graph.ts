import { forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { SimulationNodeDatum } from 'd3-force';

type Node = SimulationNodeDatum & { id: string };
type Link = { source: string; target: string; tags: string[] };

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
 * Where the notes go, in % of the view: the more tags two notes share, the closer they are.
 * d3 seeds its randomness, so the same notes get the same map on every load.
 * shortcut: titles overlap inside big clusters, add zoom when a graph gets too dense to read.
 */
export function layoutGraph(notes: { id: string }[], links: Link[]) {
  const nodes: Node[] = notes.map(({ id }) => ({ id }));
  // d3 writes into the links it gets: they are copies.
  const edges = links.map(link => ({ ...link }));

  forceSimulation(nodes)
    .force('charge', forceManyBody().strength(-300))
    .force('link', forceLink<Node, Link>(edges).id(node => node.id).distance(link => 70 / link.tags.length).strength(1))
    .force('collide', forceCollide(COLLIDE_RADIUS))
    .force('x', forceX().strength(0.04))
    .force('y', forceY().strength(0.08))
    .stop()
    .tick(TICKS);

  const x = fit(nodes.map(node => node.x!));
  const y = fit(nodes.map(node => node.y!));
  return new Map(nodes.map(node => [node.id, { x: x(node.x!), y: y(node.y!) }]));
}
