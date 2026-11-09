import assert from 'node:assert/strict';
import { test } from 'node:test';
import { MarkdownManager } from '@tiptap/markdown';
import StarterKit from '@tiptap/starter-kit';
import { BlockMath, InlineMath, readBlockMath, readInlineMath } from '../app/utils/math.ts';

const markdown = new MarkdownManager({ extensions: [StarterKit, InlineMath, BlockMath] });
const roundTrip = (source: string) => markdown.serialize(markdown.parse(source)).trim();

test('reads an inline formula at the start only', () => {
  assert.deepEqual(readInlineMath('$a_1 * b$ suite'), { type: 'inlineMath', raw: '$a_1 * b$', latex: 'a_1 * b' });
  assert.equal(readInlineMath('$x$')?.latex, 'x');
  assert.equal(readInlineMath('voir $x$'), undefined);
});

// A price is not a formula: what is not read stays the text written.
test('leaves dollars that are not a formula', () => {
  assert.equal(readInlineMath('$5 et $10'), undefined);
  assert.equal(readInlineMath('$ x$'), undefined);
  assert.equal(readInlineMath('$x $'), undefined);
  assert.equal(readInlineMath('$x$5'), undefined);
  assert.equal(readInlineMath('$$x$$'), undefined);
});

test('reads a block formula on one line or several', () => {
  assert.deepEqual(readBlockMath('$$\n\\sum_i x_i\n$$\n\nsuite'), {
    type: 'blockMath',
    raw: '$$\n\\sum_i x_i\n$$\n',
    latex: '\\sum_i x_i',
  });
  assert.equal(readBlockMath('$$x^2$$')?.latex, 'x^2');
  assert.equal(readBlockMath('$$x$$ suite'), undefined);
});

// Saved then opened again, a note keeps its text: the _ and * of a formula are not read as markdown.
test('writes the formulas back as they were written', () => {
  for (const source of [
    'Soit $a_1 * b_2$ et $x^2$.',
    '$$\n\\sum_{i=1}^n x_i\n$$',
    'Il coûte $5 et $10.',
    'Avant\n\n$$\nE = mc^2\n$$\n\nAprès',
  ]) {
    assert.equal(roundTrip(source), source);
  }
  assert.equal(markdown.parse('Soit $a_1 * b_2$').content![0]!.content![1]!.type, 'inlineMath');
});
