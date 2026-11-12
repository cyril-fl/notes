import assert from 'node:assert/strict';
import { test } from 'node:test';
import { layoutGraph, neighborhood, sharedTags } from '../app/utils/graph.ts';

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

test('the neighborhood keeps the note, its neighbors and the links between them', () => {
  const graph = {
    notes: notes('a', 'b', 'c', 'd'),
    links: [link('a', 'b'), link('c', 'a'), link('b', 'c'), link('c', 'd')],
  };
  assert.deepEqual(neighborhood(graph, 'a'), {
    notes: notes('a', 'b', 'c'),
    links: [link('a', 'b'), link('c', 'a'), link('b', 'c')],
  });
  assert.deepEqual(neighborhood(graph, 'x'), { notes: [], links: [] });
});

test('the shared tags are counted whichever side of the link the note is on', () => {
  const links = [link('a', 'b', 'go', 'web'), link('c', 'a', 'go'), link('b', 'c', 'go'), link('a', 'd')];
  assert.deepEqual(
    sharedTags(links, 'a'),
    new Map([
      ['b', 2],
      ['c', 1],
      ['d', 0],
    ]),
  );
  assert.deepEqual(sharedTags(links, 'x'), new Map());
});
