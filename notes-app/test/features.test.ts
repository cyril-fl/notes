import assert from 'node:assert/strict';
import { test } from 'node:test';
import { planOf, unlocked } from '../app/utils/features.ts';

test('a feature is for its plan and the plans after it, an off one for no one', () => {
  assert.deepEqual([unlocked('base', 'base'), unlocked('base', 'pro')], [true, true]);
  assert.deepEqual([unlocked('pro', 'base'), unlocked('pro', 'pro')], [false, true]);
  assert.deepEqual([unlocked('off', 'base'), unlocked('off', 'pro')], [false, false]);
});

test('the plan of an account is one of plans, the base one when unknown', () => {
  assert.deepEqual(
    [planOf({ plan: 'base' }), planOf({ plan: 'pro' }), planOf({ plan: 'max' })],
    ['base', 'pro', 'base'],
  );
});
