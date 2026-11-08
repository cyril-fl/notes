import assert from 'node:assert/strict';
import { test } from 'node:test';
import { countWords } from '../app/utils/words.ts';

test('counts the words, not the punctuation', () => {
  assert.equal(countWords(''), 0);
  assert.equal(countWords('Bonjour — le monde !'), 3);
  assert.equal(countWords("l'arbre\n\nest vert, 12 fois."), 5);
});
