import { forceCollide, forceLink, forceManyBody, forceSimulation, forceX, forceY } from 'd3-force';
import type { SimulationNodeDatum } from 'd3-force';

type Node = SimulationNodeDatum & { id: string };
type Note = { id: string; tags?: string[] };
type Link = { source: string; target: string };
type Edge = Link & { strength: number; distance: number };

const TICKS = 300;
const MARGIN = 8;
const COLLIDE_RADIUS = 30;
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

/**
 * Where the notes and their tags go, in % of the view: see graphEdges for the springs.
 * The pull to the center weakens as the graph grows (1/√ of its nodes), or a big one shrinks into a ball.
 * d3 seeds its randomness, so the same notes get the same map on every load. GraphView zooms by spreading them.
 */
export function layoutGraph(notes: Note[], links: Link[]) {
  const tags = new Set(notes.flatMap((note) => note.tags ?? []));
  const nodes: Node[] = [...notes.map(({ id }) => ({ id })), ...[...tags].map((tag) => ({ id: tagNode(tag) }))];
  const edges = graphEdges(notes, links);
  const gravity = Math.sqrt(10 / Math.max(nodes.length, 10));

  forceSimulation(nodes)
    .force('charge', forceManyBody().strength(-300))
    .force(
      'link',
      forceLink<Node, Edge>(edges)
        .id((node) => node.id)
        .distance((edge) => edge.distance)
        .strength((edge) => edge.strength),
    )
    .force('collide', forceCollide(COLLIDE_RADIUS))
    .force('x', forceX().strength(0.04 * gravity))
    .force('y', forceY().strength(0.08 * gravity))
    .stop()
    .tick(TICKS);

  const x = fit(nodes.map((node) => node.x!));
  const y = fit(nodes.map((node) => node.y!));
  return new Map(nodes.map((node) => [node.id, { x: x(node.x!), y: y(node.y!) }]));
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
