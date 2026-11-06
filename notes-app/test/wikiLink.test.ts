import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readWikiLink, writeWikiLink } from '../app/utils/wikiLink.ts';

const ID = '3f2a0000-0000-0000-0000-00000000abcd';

test('reads a waiting link at the start only', () => {
  assert.deepEqual(readWikiLink('[[Courses]] et la suite'), { type: 'wikiLink', raw: '[[Courses]]', target: 'Courses', id: null });
  assert.deepEqual(readWikiLink('[[ /A/Idées ]]'), { type: 'wikiLink', raw: '[[ /A/Idées ]]', target: '/A/Idées', id: null });
  assert.equal(readWikiLink('voir [[Courses]]'), undefined);
});

test('reads a resolved link with its id', () => {
  assert.deepEqual(readWikiLink(`[Courses](note:${ID.toUpperCase()}) suite`), { type: 'wikiLink', raw: `[Courses](note:${ID.toUpperCase()})`, target: 'Courses', id: ID });
});

test('leaves the other links and brackets to markdown', () => {
  for (const src of ['[[]]', '[[  ]]', '[[a\nb]]', '[[a[b]]', '[a](b)', '[a](note:nope)', '[seul]']) {
    assert.equal(readWikiLink(src), undefined, src);
  }
});

test('writes by id once resolved, by title until then', () => {
  assert.equal(writeWikiLink('Courses [x]', ID), `[Courses x](note:${ID})`);
  assert.equal(writeWikiLink('Courses', null), '[[Courses]]');
});
