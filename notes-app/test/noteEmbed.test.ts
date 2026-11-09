import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readNoteEmbed, writeNoteEmbed } from '../app/utils/noteEmbed.ts';

const ID = '3f2a0000-0000-0000-0000-00000000abcd';

test('reads an embed waiting or resolved, at the start only', () => {
  assert.deepEqual(readNoteEmbed('![[Courses]] et la suite'), {
    type: 'noteEmbed',
    raw: '![[Courses]]',
    target: 'Courses',
    id: null,
    heading: null,
  });
  assert.deepEqual(readNoteEmbed(`![Courses](note:${ID})`), {
    type: 'noteEmbed',
    raw: `![Courses](note:${ID})`,
    target: 'Courses',
    id: ID,
    heading: null,
  });
  assert.equal(readNoteEmbed('[[Courses]]'), undefined);
  assert.equal(readNoteEmbed('![photo](/api/files/1)'), undefined);
});

test('writes an embed as a link after a !', () => {
  assert.equal(writeNoteEmbed('Courses', null), '![[Courses]]');
  assert.equal(writeNoteEmbed('Courses', ID), `![Courses](note:${ID})`);
});
