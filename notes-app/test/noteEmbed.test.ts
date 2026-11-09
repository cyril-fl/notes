import assert from 'node:assert/strict';
import { test } from 'node:test';
import { readNoteEmbed, sectionOf, withoutTitle, writeNoteEmbed } from '../app/utils/noteEmbed.ts';

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

test('drops the title heading an embedded note starts with, not a later one', () => {
  assert.equal(withoutTitle('# Courses\n\n- lait\n\n## Fin'), '- lait\n\n## Fin');
  assert.equal(withoutTitle('texte\n\n# Plus bas'), 'texte\n\n# Plus bas');
});

test('keeps the section under a heading, down to the next one of its level or above', () => {
  const note =
    '# Courses\n\nintro\n\n## Lait\n\nentier\n\n### Marque\n\nbio\n\n```\n# pas un titre\n```\n\n## Pain\n\nbaguette';
  assert.equal(sectionOf(note, 'lait'), '## Lait\n\nentier\n\n### Marque\n\nbio\n\n```\n# pas un titre\n```');
  assert.equal(sectionOf(note, 'Pain'), '## Pain\n\nbaguette');
  assert.equal(sectionOf(note, 'Absent'), note);
});
