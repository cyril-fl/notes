import assert from 'node:assert/strict';
import { test } from 'node:test';
import { retitle } from '../app/utils/title.ts';

test('retitle replaces the first heading outside code, else adds one', () => {
  assert.equal(retitle('intro\n## Old\n# Next', 'New'), 'intro\n## New\n# Next');
  assert.equal(retitle('```\n# comment\n```\n# Old', 'New'), '```\n# comment\n```\n# New');
  assert.equal(retitle('#tag only\nfirst line', 'New'), '# New\n\n#tag only\nfirst line');
  assert.equal(retitle('', 'New'), '# New');
});
