import assert from 'node:assert/strict';
import { test } from 'node:test';
import { tabAfterClose } from '../app/utils/tabs.ts';

test('closing a tab opens the one before it, else the one after it', () => {
  assert.equal(tabAfterClose(['a', 'b', 'c'], 'b'), 'a');
  assert.equal(tabAfterClose(['a', 'b', 'c'], 'c'), 'b');
  assert.equal(tabAfterClose(['a', 'b', 'c'], 'a'), 'b');
});

test('the last tab closed leaves none, an unknown one none either', () => {
  assert.equal(tabAfterClose(['a'], 'a'), undefined);
  assert.equal(tabAfterClose([], 'a'), undefined);
  assert.equal(tabAfterClose(['a', 'b'], 'c'), undefined);
});
