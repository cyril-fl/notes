import assert from 'node:assert/strict';
import { test } from 'node:test';
import { layoutGraph, linkTargets, neighborGraph } from '../app/utils/graph.ts';

const notes = (...ids: string[]) => ids.map((id) => ({ id }));
const link = (source: string, target: string, ...tags: string[]) => ({ source, target, tags });

function distance(positions: ReturnType<typeof layoutGraph>, a: string, b: string) {
  const from = positions.get(a)!;
  const to = positions.get(b)!;
  return Math.hypot(from.x - to.x, from.y - to.y);
}

test('no note, no position', () => {
  assert.equal(layoutGraph([], []).size, 0);
});

test('one note sits in the center', () => {
  assert.deepEqual(layoutGraph(notes('a'), []).get('a'), { x: 50, y: 50 });
});

test('notes sharing tags gather, away from the others', () => {
  const positions = layoutGraph(notes('a', 'b', 'c', 'd', 'e', 'f'), [
    link('a', 'b', 'go'),
    link('b', 'c', 'go'),
    link('a', 'c', 'go'),
    link('d', 'e', 'cuisine'),
    link('e', 'f', 'cuisine'),
    link('d', 'f', 'cuisine'),
  ]);
  const inside = Math.max(distance(positions, 'a', 'b'), distance(positions, 'b', 'c'), distance(positions, 'a', 'c'));
  const across = Math.min(
    ...['d', 'e', 'f'].flatMap((other) => ['a', 'b', 'c'].map((id) => distance(positions, id, other))),
  );
  assert.ok(inside < across, `${inside} inside, ${across} across`);
});

test('more shared tags bring notes closer', () => {
  const positions = layoutGraph(notes('a', 'b', 'c'), [link('a', 'b', 'x', 'y', 'z'), link('b', 'c', 'x')]);
  assert.ok(distance(positions, 'a', 'b') < distance(positions, 'b', 'c'));
});

test('every note stays in the view, with its links untouched', () => {
  const ids = Array.from({ length: 80 }, (_, i) => `n${i}`);
  const links = ids.slice(1).map((id) => link('n0', id, 'go'));
  const positions = layoutGraph(notes(...ids), links);
  assert.equal(positions.size, 80);
  for (const { x, y } of positions.values()) assert.ok(x >= 8 && x <= 92 && y >= 8 && y <= 92, `${x}, ${y}`);
  assert.equal(links[0]!.source, 'n0');
});

test('the same notes get the same map', () => {
  const graph = [notes('a', 'b', 'c'), [link('a', 'b', 'go')]] as const;
  assert.deepEqual(layoutGraph(...graph), layoutGraph(...graph));
});

test('a [[link]] alone pulls notes together like a shared tag', () => {
  const positions = layoutGraph(notes('a', 'b', 'c'), [{ source: 'a', target: 'b', tags: [], linked: true }]);
  for (const id of ['a', 'b', 'c']) {
    const { x, y } = positions.get(id)!;
    assert.ok(Number.isFinite(x) && Number.isFinite(y), `${id} at ${x}, ${y}`);
  }
  assert.ok(distance(positions, 'a', 'b') < distance(positions, 'a', 'c'));
});

test('the graph of the neighbors joins the note to each, a [[link]] in either way merged with the shared tags', () => {
  const near = (id: string) => ({ id, title: id.toUpperCase(), path: `/${id}` });
  const graph = neighborGraph({
    note: { item: { id: 'a', title: 'A' }, path: '/a' },
    neighbors: [
      { ...near('b'), tags: ['go', 'web'] },
      { ...near('c'), tags: ['go'] },
    ],
    links: [near('c'), near('a')],
    backlinks: [near('d')],
  });
  assert.deepEqual(graph.notes, ['a', 'b', 'c', 'd'].map(near));
  assert.deepEqual(graph.links, [
    { source: 'a', target: 'b', tags: ['go', 'web'], linked: false },
    { source: 'a', target: 'c', tags: ['go'], linked: true },
    { source: 'a', target: 'd', tags: [], linked: true },
  ]);
});

test('the [[ menu offers the other named notes, the closest ones first for a Pro account', () => {
  const notes = [
    { id: 'a', title: 'A' },
    { id: 'b', title: 'B' },
    { id: 'c', title: 'C' },
    { id: 'd', title: 'D' },
    { id: 'e', title: '' },
  ];
  const neighbors = [
    { id: 'c', shared: 2 },
    { id: 'b', shared: 1 },
  ];
  const titles = (id: string | undefined, near?: typeof neighbors) =>
    linkTargets(notes, id, near).map((note) => note.title);
  assert.deepEqual(titles('a', neighbors), ['C', 'B', 'D']);
  assert.deepEqual(titles('a'), ['B', 'C', 'D']);
  assert.deepEqual(titles(undefined), ['A', 'B', 'C', 'D']);
  assert.deepEqual(linkTargets(null, 'a', neighbors), []);
});
