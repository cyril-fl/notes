import { forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { SimulationNodeDatum } from 'd3-force';

type Node = SimulationNodeDatum & { id: string };
type Note = { id: string; tags?: string[] };
type Link = { source: string; target: string };
type Edge = Link & { strength: number; distance: number };

// Measured on a synthetic account: half the ticks of d3's full cooling, the same clusters, half the time.
const TICKS = 150;
const MARGIN = 8;
const COLLIDE_RADIUS = 30;
// Collisions keep the titles of a small graph apart, all shown at once; past it they cost more than every other force
// and the zoom spreads the titles anyway.
const COLLIDE_LIMIT = 500;
// Barnes-Hut approximation of the repulsion, coarser than d3's 0.9: a third faster, the same clusters.
const THETA = 1.5;
const LINK_DISTANCE = 2 * COLLIDE_RADIUS;
// A tag on every note says nothing of them, but must still hold them.
const MIN_WEIGHT = 0.05;
// In % of the view per unit of the simulation: keeps a small graph from spreading to the corners.
const MAX_SCALE = 0.35;

/** The node of a tag: notes have uuids, no clash. */
export const tagNode = (tag: string) => `#${tag}`;

/** Scales the simulated values into the view, around its center. */
function fit(values: number[]) {
  const min = Math.min(...values);
  const range = Math.max(...values) - min;
  const scale = range ? Math.min((100 - 2 * MARGIN) / range, MAX_SCALE) : 0;
  return (value: number) => 50 + (value - min - range / 2) * scale;
}

/**
 * The springs of the graph, a note to each of its tags and to the notes of its [[links]].
 * - A tag pulls by how rare it is among the n notes, log(n / its notes) / log(n): a tag on half of them says little.
 *   Shared among the tags of the note, so a note with many tags is not pulled harder than one with a single tag.
 * - Its notes sit at √(its notes) collide radii: the room a disk of them needs, so a big tag spreads instead of piling up.
 * - A [[link]] is a choice, not a vote: short and strong, weakened only on a note with many of them.
 */
export function graphEdges(notes: Note[], links: Link[]) {
  const carriers = new Map<string, number>();
  for (const note of notes) for (const tag of note.tags ?? []) carriers.set(tag, (carriers.get(tag) ?? 0) + 1);
  const linked = new Map<string, number>();
  for (const { source, target } of links) {
    linked.set(source, (linked.get(source) ?? 0) + 1);
    linked.set(target, (linked.get(target) ?? 0) + 1);
  }

  const n = notes.length;
  const weight = (count: number) => (n > 1 ? Math.max(Math.log(n / count) / Math.log(n), MIN_WEIGHT) : 1);
  const edges: Edge[] = notes.flatMap(({ id, tags = [] }) =>
    tags.map((tag) => {
      const count = carriers.get(tag)!;
      return {
        source: id,
        target: tagNode(tag),
        strength: weight(count) / tags.length,
        distance: Math.max(COLLIDE_RADIUS * Math.sqrt(count), LINK_DISTANCE),
      };
    }),
  );
  for (const link of links)
    edges.push({
      ...link,
      strength: 1 / Math.min(linked.get(link.source)!, linked.get(link.target)!),
      distance: LINK_DISTANCE,
    });
  return edges;
}

type Point = { x: number; y: number };
// From a saved layout, only what changed settles: a little heat, a few ticks.
const WARM_ALPHA = 0.1;
const WARM_TICKS = 40;

/**
 * Where the notes and their tags go, in the simulation's units: see graphEdges for the springs.
 * The pull to the center weakens as the graph grows (1/√ of its nodes), or a big one shrinks into a ball.
 * seed: a layout saved before; the nodes it lacks start next to a neighbor it has, and the whole only settles a little.
 * d3 seeds its randomness, so the same notes get the same map on every load.
 */
export function simulateGraph(notes: Note[], links: Link[], seed?: Map<string, Point>) {
  const tags = new Set(notes.flatMap((note) => note.tags ?? []));
  const nodes: Node[] = [...notes.map(({ id }) => ({ id })), ...[...tags].map((tag) => ({ id: tagNode(tag) }))];
  const edges = graphEdges(notes, links);
  const gravity = Math.sqrt(10 / Math.max(nodes.length, 10));

  if (seed) {
    const near = new Map<string, Point>();
    for (const { source, target } of edges) {
      if (seed.has(source) && !near.has(target)) near.set(target, seed.get(source)!);
      if (seed.has(target) && !near.has(source)) near.set(source, seed.get(target)!);
    }
    for (const node of nodes) {
      const at = seed.get(node.id) ?? near.get(node.id);
      // Beside its neighbor, not on it: two nodes at the same point have no direction to push apart.
      if (at) [node.x, node.y] = seed.has(node.id) ? [at.x, at.y] : [at.x + COLLIDE_RADIUS, at.y + COLLIDE_RADIUS];
    }
  }

  const simulation = forceSimulation(nodes)
    .force('charge', forceManyBody().strength(-300).theta(THETA))
    .force(
      'link',
      forceLink<Node, Edge>(edges)
        .id((node) => node.id)
        .distance((edge) => edge.distance)
        .strength((edge) => edge.strength),
    )
    .force('collide', nodes.length <= COLLIDE_LIMIT ? forceCollide(COLLIDE_RADIUS) : null)
    .force('x', forceX().strength(0.04 * gravity))
    .force('y', forceY().strength(0.08 * gravity))
    .stop();
  if (seed) simulation.alpha(WARM_ALPHA);
  simulation.tick(seed ? WARM_TICKS : TICKS);

  return new Map(nodes.map((node) => [node.id, { x: node.x!, y: node.y! }]));
}

/** A simulated layout in % of the view. GraphView zooms by spreading it. */
export function fitLayout(positions: Map<string, Point>) {
  const x = fit([...positions.values()].map((at) => at.x));
  const y = fit([...positions.values()].map((at) => at.y));
  return new Map([...positions].map(([id, at]) => [id, { x: x(at.x), y: y(at.y) }]));
}

/** Where the notes and their tags go, in % of the view, from scratch. */
export function layoutGraph(notes: Note[], links: Link[]) {
  return fitLayout(simulateGraph(notes, links));
}

/** A fingerprint of the notes, their tags and links (FNV-1a): a saved layout serves as is only for the same graph. */
export function graphSignature(notes: Note[], links: Link[]) {
  let hash = 0x811c9dc5;
  const add = (text: string) => {
    for (let i = 0; i < text.length; i++) hash = Math.imul(hash ^ text.charCodeAt(i), 0x01000193);
  };
  for (const { id, tags = [] } of notes) add(`${id}:${tags.join(',')};`);
  for (const { source, target } of links) add(`${source}>${target};`);
  return `${notes.length}:${links.length}:${(hash >>> 0).toString(36)}`;
}

// A tag on more than this share of the notes says little of each: the graph can leave it out.
export const WIDE_TAG_SHARE = 0.2;

/** The tags on more than WIDE_TAG_SHARE of the notes, biggest first, and the graph without them. */
export function withoutWideTags<G extends { notes: Note[] }>(graph: G) {
  const counts = new Map<string, number>();
  for (const note of graph.notes) for (const tag of note.tags ?? []) counts.set(tag, (counts.get(tag) ?? 0) + 1);
  const wide = [...counts]
    .filter(([, count]) => count > graph.notes.length * WIDE_TAG_SHARE)
    .sort((a, b) => b[1] - a[1])
    .map(([tag]) => tag);
  const hidden = new Set(wide);
  const notes = graph.notes.map((note) => ({ ...note, tags: note.tags?.filter((tag) => !hidden.has(tag)) }));
  return { wide, graph: { ...graph, notes } };
}

type Near = { id: string; title: string; path: string };

/**
 * The graph around one note, from its neighbors: the note with its tags, the notes sharing a tag with it (with those tags only),
 * the notes of its [[links]] either way, and those links.
 */
export function neighborGraph(near: {
  note: { item: { id: string; title: string }; path: string; tags?: string[] };
  neighbors: (Near & { tags: string[] })[];
  links: Near[];
  backlinks: Near[];
}) {
  const id = near.note.item.id;
  const notes = new Map<string, Near & { tags?: string[] }>([
    [id, { id, title: near.note.item.title, path: near.note.path, tags: near.note.tags ?? [] }],
  ]);
  for (const note of near.neighbors)
    notes.set(note.id, { id: note.id, title: note.title, path: note.path, tags: note.tags });
  const links = new Map<string, Link>();
  for (const note of [...near.links, ...near.backlinks]) {
    if (note.id === id) continue;
    if (!notes.has(note.id)) notes.set(note.id, { id: note.id, title: note.title, path: note.path });
    links.set(note.id, { source: id, target: note.id });
  }
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
