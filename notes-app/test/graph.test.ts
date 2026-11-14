import assert from 'node:assert/strict';
import { test } from 'node:test';
import { linkTargets } from '../app/utils/graph.ts';

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
