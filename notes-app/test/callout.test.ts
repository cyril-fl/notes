import assert from 'node:assert/strict';
import { test } from 'node:test';
import { getSchema } from '@tiptap/core';
import { MarkdownManager } from '@tiptap/markdown';
import StarterKit from '@tiptap/starter-kit';
import { Callout } from '../app/utils/callout.ts';

const markdown = new MarkdownManager({ extensions: [StarterKit.configure({ blockquote: false }), Callout] });
const roundTrip = (source: string) => markdown.serialize(markdown.parse(source)).trim();

test('reads the type of a callout and writes it back, with or without a title', () => {
  const doc = markdown.parse('> [!warning]\n> Attention *ici*');
  assert.equal(doc.content![0]!.attrs!.callout, 'warning');
  assert.equal(doc.content![0]!.content![0]!.content![0]!.text, 'Attention ');

  for (const source of ['> [!warning]\n> Attention *ici*', '> [!NOTE] Titre\n> corps', '> [!danger] Seul']) {
    assert.equal(roundTrip(source), source);
  }
  // The same callout for Obsidian and GitHub, without the empty line.
  assert.equal(roundTrip('> [!tip]\n>\n> second'), '> [!tip]\n> second');
});

test('names a callout in the given language, an unknown type as written', () => {
  const schema = getSchema([
    StarterKit.configure({ blockquote: false }),
    Callout.configure({ labels: { tip: 'Astuce' } }),
  ]);
  const attrs = (callout: string) =>
    (schema.nodes.blockquote!.spec.toDOM!(schema.nodes.blockquote!.create({ callout })) as [string, object])[1];
  assert.deepEqual(attrs('TIP'), { 'data-callout': 'TIP', 'data-label': 'Astuce' });
  assert.deepEqual(attrs('info'), { 'data-callout': 'info', 'data-label': 'info' });
});

test('leaves a plain quote and a marker that is not at its start as they are', () => {
  assert.equal(markdown.parse('> citation').content![0]!.attrs?.callout, undefined);
  assert.equal(markdown.parse('> **[!note]** gras').content![0]!.attrs?.callout, undefined);
  assert.equal(roundTrip('> citation\n> suite'), '> citation\n> suite');
});
