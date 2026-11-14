import assert from 'node:assert/strict';
import { test } from 'node:test';
import { createGlobalState } from '@vueuse/core';
import { nextTick, reactive, ref, watch } from 'vue';
import { canDropInto, selectRange, toggleSelected } from '../app/utils/selection.ts';

test('a ⌘-click adds an item, a second one takes it out', () => {
  assert.deepEqual(toggleSelected([], 'a'), ['a']);
  assert.deepEqual(toggleSelected(['a', 'b'], 'a'), ['b']);
});

test('a ⇧-click adds the range in either direction, without doubles', () => {
  const ids = ['a', 'b', 'c', 'd'];
  assert.deepEqual(selectRange([], ids, 1, 3), ['b', 'c', 'd']);
  assert.deepEqual(selectRange(['c'], ids, 2, 0), ['c', 'a', 'b']);
});

test('a ⇧-click range skips the items not read yet', () => {
  assert.deepEqual(selectRange([], ['a', undefined, undefined, 'd'], 0, 3), ['a', 'd']);
});

const drag = (itemId: string, parentId: string | null, itemIds = [itemId]) => ({ itemId, parentId, itemIds });

test('items drop into another folder or the root', () => {
  assert.equal(canDropInto(drag('n', 'f1'), 'f2'), true);
  assert.equal(canDropInto(drag('n', 'f1'), null), true);
});

test('items do not drop where they already are', () => {
  assert.equal(canDropInto(drag('n', 'f1'), 'f1'), false);
  assert.equal(canDropInto(drag('n', null), null), false);
});

test('a folder does not drop into itself, alone or with the selection', () => {
  assert.equal(canDropInto(drag('f1', null), 'f1'), false);
  assert.equal(canDropInto(drag('n', null, ['n', 'f1']), 'f1'), false);
});

test('only an item drops: not a file or a link from elsewhere', () => {
  assert.equal(canDropInto({}, 'f1'), false);
});

test('the selection empties when the page changes', async () => {
  const route = reactive({ fullPath: '/folders/a' });
  Object.assign(globalThis, { createGlobalState, ref, watch, useRoute: () => route });
  const { useSelection } = await import('../app/composables/useSelection.ts');

  const selection = useSelection();
  selection.value = ['a', 'b'];
  assert.equal(useSelection(), selection, 'one selection for the whole app');
  route.fullPath = '/folders/b';
  await nextTick();
  assert.deepEqual(selection.value, []);
});
