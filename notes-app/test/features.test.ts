import assert from 'node:assert/strict';
import { test } from 'node:test';
import { unlocked } from '../app/utils/features.ts';

test('a base feature is for everyone, a pro one with the Pro mode, an off one for no one', () => {
  assert.deepEqual([unlocked('base', false), unlocked('base', true)], [true, true]);
  assert.deepEqual([unlocked('pro', false), unlocked('pro', true)], [false, true]);
  assert.deepEqual([unlocked('off', false), unlocked('off', true)], [false, false]);
});
