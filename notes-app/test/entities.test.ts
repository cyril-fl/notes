import assert from 'node:assert/strict';
import { test } from 'node:test';
import { decodeEntities } from '../app/utils/entities.ts';

test('decodes & < > in text', () => {
  assert.equal(decodeEntities('a &gt;&gt; b &amp; c &lt; d'), 'a >> b & c < d');
});

test('a typed entity comes back as typed', () => {
  assert.equal(decodeEntities('&amp;lt; et &amp;amp;'), '&lt; et &amp;');
});

test('leaves code as it is', () => {
  assert.equal(decodeEntities('`&lt;` &lt;'), '`&lt;` <');
  assert.equal(decodeEntities('```\n&amp;\n```\n&amp;'), '```\n&amp;\n```\n&');
  assert.equal(decodeEntities('~~~go\na &lt; b\n~~~'), '~~~go\na &lt; b\n~~~');
});
