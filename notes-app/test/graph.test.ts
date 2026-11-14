import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  fitLayout,
  graphEdges,
  graphSignature,
  layoutGraph,
  linkTargets,
  neighborGraph,
  simulateGraph,
  tagNode,
} from '../app/utils/graph.ts';

const note = (id: string, ...tags: string[]) => ({ id, tags });
const link = (source: string, target: string) => ({ source, target });

function distance(positions: ReturnType<typeof layoutGraph>, a: string, b: string) {
  const from = positions.get(a)!;
  const to = positions.get(b)!;
  return Math.hypot(from.x - to.x, from.y - to.y);
}

test('no note, no position', () => {
  assert.equal(layoutGraph([], []).size, 0);
});

test('one note sits in the center', () => {
  assert.deepEqual(layoutGraph([note('a')], []).get('a'), { x: 50, y: 50 });
});

test('the tags are nodes, joined to their notes', () => {
  const positions = layoutGraph([note('a', 'go'), note('b', 'go')], []);
  assert.deepEqual([...positions.keys()], ['a', 'b', tagNode('go')]);
});

test('a rare tag pulls harder than a common one, shared among the tags of the note', () => {
  const notes = [
    note('a', 'rare', 'common'),
    note('b', 'rare', 'common'),
    ...'cdefgh'.split('').map((id) => note(id, 'common')),
  ];
  const edges = graphEdges(notes, []);
  const pull = (id: string, tag: string) =>
    edges.find((edge) => edge.source === id && edge.target === tagNode(tag))!.strength;
  assert.ok(pull('a', 'rare') > pull('a', 'common'));
  // c has the common tag alone: the whole of its weight, where a splits it in two.
  assert.equal(pull('c', 'common'), 2 * pull('a', 'common'));
});

test('a tag on every note still holds them, a lone note is pulled fully', () => {
  assert.ok(graphEdges([note('a', 'all'), note('b', 'all')], [])[0]!.strength > 0);
  assert.equal(graphEdges([note('a', 'go')], [])[0]!.strength, 1);
});

test('a big tag keeps its notes further away than a small one', () => {
  const notes = Array.from({ length: 30 }, (_, i) => note(`n${i}`, 'big', ...(i < 2 ? ['small'] : [])));
  const edges = graphEdges(notes, []);
  const far = (tag: string) => edges.find((edge) => edge.target === tagNode(tag))!.distance;
  assert.ok(far('big') > far('small'));
});

test('notes of a small tag, or joined by a [[link]], gather closer than notes sharing only a big one', () => {
  const notes = Array.from({ length: 40 }, (_, i) => note(`n${i}`, 'big'));
  notes[0]!.tags.push('small');
  notes[1]!.tags.push('small');
  const positions = layoutGraph(notes, [link('n2', 'n3')]);
  const spread = distance(positions, 'n10', 'n20');
  assert.ok(distance(positions, 'n0', 'n1') < spread, `small ${distance(positions, 'n0', 'n1')}, big ${spread}`);
  assert.ok(distance(positions, 'n2', 'n3') < spread, `linked ${distance(positions, 'n2', 'n3')}, big ${spread}`);
});

test('notes sharing a tag gather around it, away from the others', () => {
  const groups = { go: ['a', 'b', 'c'], cuisine: ['d', 'e', 'f'] };
  const positions = layoutGraph(
    Object.entries(groups).flatMap(([tag, ids]) => ids.map((id) => note(id, tag))),
    [],
  );
  for (const [tag, ids] of Object.entries(groups)) {
    const other = tag === 'go' ? 'cuisine' : 'go';
    for (const id of ids)
      assert.ok(
        distance(positions, id, tagNode(tag)) < distance(positions, id, tagNode(other)),
        `${id} nearer ${other}`,
      );
  }
});

test('a [[link]] alone pulls notes together', () => {
  const positions = layoutGraph([note('a'), note('b'), note('c')], [link('a', 'b')]);
  for (const id of ['a', 'b', 'c']) {
    const { x, y } = positions.get(id)!;
    assert.ok(Number.isFinite(x) && Number.isFinite(y), `${id} at ${x}, ${y}`);
  }
  assert.ok(distance(positions, 'a', 'b') < distance(positions, 'a', 'c'));
});

test('every note stays in the view, with its links untouched', () => {
  const notes = Array.from({ length: 80 }, (_, i) => note(`n${i}`, 'go'));
  const links = [link('n0', 'n1')];
  const positions = layoutGraph(notes, links);
  assert.equal(positions.size, 81);
  for (const { x, y } of positions.values()) assert.ok(x >= 8 && x <= 92 && y >= 8 && y <= 92, `${x}, ${y}`);
  assert.deepEqual(links[0], link('n0', 'n1'));
});

test('the same notes get the same map', () => {
  const graph = [[note('a', 'go'), note('b', 'go'), note('c')], [link('a', 'c')]] as const;
  assert.deepEqual(layoutGraph(...graph), layoutGraph(...graph));
});

test('the signature follows the notes, their tags and links', () => {
  const notes = [note('a', 'go'), note('b')];
  assert.equal(graphSignature(notes, []), graphSignature([note('a', 'go'), note('b')], []));
  assert.notEqual(graphSignature(notes, []), graphSignature([note('a', 'web'), note('b')], []));
  assert.notEqual(graphSignature(notes, []), graphSignature(notes, [link('a', 'b')]));
});

test('from a saved layout, the notes stay in place and a new one joins its tag', () => {
  const groups = { go: ['a', 'b', 'c'], cuisine: ['d', 'e', 'f'] };
  const notes = Object.entries(groups).flatMap(([tag, ids]) => ids.map((id) => note(id, tag)));
  const saved = simulateGraph(notes, []);
  const warm = simulateGraph([...notes, note('g', 'cuisine')], [], saved);
  const before = fitLayout(saved);
  const after = fitLayout(warm);
  const at = (positions: typeof warm, id: string) => positions.get(id)!;
  const gap = (from: { x: number; y: number }, to: { x: number; y: number }) =>
    Math.hypot(from.x - to.x, from.y - to.y);
  for (const id of 'abcdef') assert.ok(gap(at(before, id), at(after, id)) < 10, `${id} moved`);
  assert.ok(gap(at(warm, 'g'), at(warm, tagNode('cuisine'))) < gap(at(warm, 'g'), at(warm, tagNode('go'))));
});

test('the graph of the neighbors keeps the note with its tags, the neighbors with the shared ones, and the [[links]] either way', () => {
  const near = (id: string) => ({ id, title: id.toUpperCase(), path: `/${id}` });
  const graph = neighborGraph({
    note: { item: { id: 'a', title: 'A' }, path: '/a', tags: ['go', 'web', 'own'] },
    neighbors: [
      { ...near('b'), tags: ['go', 'web'] },
      { ...near('c'), tags: ['go'] },
    ],
    links: [near('c'), near('a')],
    backlinks: [near('d')],
  });
  assert.deepEqual(graph.notes, [
    { ...near('a'), tags: ['go', 'web', 'own'] },
    { ...near('b'), tags: ['go', 'web'] },
    { ...near('c'), tags: ['go'] },
    near('d'),
  ]);
  assert.deepEqual(graph.links, [link('a', 'c'), link('a', 'd')]);
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
