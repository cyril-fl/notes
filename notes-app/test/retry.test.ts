import assert from 'node:assert/strict';
import { test } from 'node:test';
import type { FetchError } from 'ofetch';
import { retryWait } from '../app/utils/retry.ts';

const after = (seconds: string) => ({ response: { headers: new Headers({ 'Retry-After': seconds }) } }) as FetchError;

test('says the wait in s, min or h', () => {
  assert.equal(retryWait(after('30')), '30 s');
  assert.equal(retryWait(after('60')), '1 min');
  assert.equal(retryWait(after('290')), '5 min');
  assert.equal(retryWait(after('3600')), '1 h');
});
