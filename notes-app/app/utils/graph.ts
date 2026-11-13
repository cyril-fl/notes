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

type Near = { id: string; title: string; path: string };

/**
 * The graph around one note, from its neighbors: the note, the notes it shares a [[link]] or a tag with, and its own links only.
 * The links between the neighbors are left out: a tag on hundreds of notes makes tens of thousands of them.
 */
export function neighborGraph(near: {
  note: { item: { id: string; title: string }; path: string };
  neighbors: (Near & { tags: string[] })[];
  links: Near[];
  backlinks: Near[];
}) {
  const id = near.note.item.id;
  const notes = new Map<string, Near>([[id, { id, title: near.note.item.title, path: near.note.path }]]);
  const links = new Map<string, Required<Link>>();
  const add = (note: Near, tags: string[], linked: boolean) => {
    if (note.id === id) return;
    notes.set(note.id, { id: note.id, title: note.title, path: note.path });
    links.set(note.id, { source: id, target: note.id, tags, linked });
  };
  for (const note of near.neighbors) add(note, note.tags, false);
  for (const note of [...near.links, ...near.backlinks]) add(note, links.get(note.id)?.tags ?? [], true);
  return { notes: [...notes.values()], links: [...links.values()] };
}

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
