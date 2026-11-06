import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readWikiLink } from '../app/utils/wikiLink.ts';

test('reads the link at the start only', () => {
  assert.deepEqual(readWikiLink('[[Courses]] et la suite'), { type: 'wikiLink', raw: '[[Courses]]', target: 'Courses' });
  assert.deepEqual(readWikiLink('[[ /A/Idées ]]'), { type: 'wikiLink', raw: '[[ /A/Idées ]]', target: '/A/Idées' });
  assert.equal(readWikiLink('voir [[Courses]]'), undefined);
});

test('refuses empty, multi-line or nested links', () => {
  for (const src of ['[[]]', '[[  ]]', '[[a\nb]]', '[[a[b]]', '[a](b)']) {
    assert.equal(readWikiLink(src), undefined, src);
  }
});
