import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  headingsOf,
  linkedNotes,
  readWikiLink,
  resolvedLinks,
  sameHeading,
  writeWikiLink,
} from '../app/utils/wikiLink.ts';

const ID = '3f2a0000-0000-0000-0000-00000000abcd';

test('reads a waiting link at the start only', () => {
  assert.deepEqual(readWikiLink('[[Courses]] et la suite'), {
    type: 'wikiLink',
    raw: '[[Courses]]',
    target: 'Courses',
    id: null,
    heading: null,
  });
  assert.deepEqual(readWikiLink('[[ /A/Idées ]]'), {
    type: 'wikiLink',
    raw: '[[ /A/Idées ]]',
    target: '/A/Idées',
    id: null,
    heading: null,
  });
  assert.equal(readWikiLink('voir [[Courses]]'), undefined);
});

test('reads a resolved link with its id', () => {
  assert.deepEqual(readWikiLink(`[Courses](note:${ID.toUpperCase()}) suite`), {
    type: 'wikiLink',
    raw: `[Courses](note:${ID.toUpperCase()})`,
    target: 'Courses',
    id: ID,
    heading: null,
  });
  assert.deepEqual(readWikiLink(`[Courses#Lait bio](note:${ID}#Lait bio)`), {
    type: 'wikiLink',
    raw: `[Courses#Lait bio](note:${ID}#Lait bio)`,
    target: 'Courses#Lait bio',
    id: ID,
    heading: 'Lait bio',
  });
});

test('leaves the other links and brackets to markdown', () => {
  for (const src of ['[[]]', '[[  ]]', '[[a\nb]]', '[[a[b]]', '[a](b)', '[a](note:nope)', '[seul]']) {
    assert.equal(readWikiLink(src), undefined, src);
  }
});

test('writes by id once resolved, by title until then', () => {
  assert.equal(writeWikiLink('Courses [x]', ID), `[Courses x](note:${ID})`);
  assert.equal(writeWikiLink('Courses', null), '[[Courses]]');
  assert.equal(writeWikiLink('Courses#Lait (bio)', ID, 'Lait (bio)'), `[Courses#Lait (bio)](note:${ID}#Lait bio)`);
  assert.equal(writeWikiLink('Courses#Lait', null, 'Lait'), '[[Courses#Lait]]');
});

test('pairs the links the API resolved with the targets typed', () => {
  const sent = '[[Autre]] puis [[/A/Idées]] et [[Inconnu]] [[autre#Fin]]';
  const saved = `[Autre](note:${ID}) puis [Idées](note:${ID.toUpperCase()}) et [[Inconnu]] [Autre#Fin](note:${ID}#Fin)`;
  assert.deepEqual(
    [...resolvedLinks(sent, saved)],
    [
      ['autre', { target: 'Autre', id: ID, heading: null }],
      ['/a/idées', { target: 'Idées', id: ID, heading: null }],
      ['autre#fin', { target: 'Autre#Fin', id: ID, heading: 'Fin' }],
    ],
  );
  assert.equal(resolvedLinks('[[Autre]]', 'texte').size, 0);
});

test('the notes a text links to, by id, sorted and once each', () => {
  const a = '0e6c1f5e-1b2a-4c3d-8e9f-0a1b2c3d4e5f';
  const b = '9F8E7D6C-5B4A-4321-8765-0FEDCBA98765';
  assert.equal(
    linkedNotes(`[B](note:${b}) and [A](note:${a}), [[Waiting]], [A again](note:${a})`),
    `${a},${b.toLowerCase()}`,
  );
  assert.equal(linkedNotes('no link'), '');
});

test('the headings of a text, not those of its code', () => {
  assert.deepEqual(headingsOf('# Un\ntexte\n## Deux  \n```sh\n# commentaire\n```\n#tag\n###### Six'), [
    'Un',
    'Deux',
    'Six',
  ]);
  assert.ok(sameHeading('Lait bio', ' lait (bio) '));
  assert.ok(!sameHeading('Lait', 'Lait bio'));
});
