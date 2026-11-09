import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readFileSync } from 'node:fs';
import { listedFeatures, planOf, unlocked } from '../app/utils/features.ts';

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

test('each feature of the comparison has its label in French and English', () => {
  for (const locale of ['fr', 'en']) {
    const labels = JSON.parse(readFileSync(new URL(`../i18n/locales/${locale}.json`, import.meta.url), 'utf8'));
    for (const { group, features: listed } of listedFeatures) {
      assert.ok(labels.settings.featureGroups[group], `${locale}: ${group}`);
      for (const feature of listed) assert.ok(labels.features[feature], `${locale}: ${feature}`);
    }
  }
});
