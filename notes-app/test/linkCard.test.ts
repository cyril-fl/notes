import assert from 'node:assert/strict';
import { test } from 'node:test';
import { DEFAULT_LINK_SIZE, isWebLink, linkContent, readLinkCard, writeLinkCard } from '../app/utils/linkCard.ts';

test('reads a mini or large card at the start only', () => {
  assert.deepEqual(readLinkCard('[Nuxt](https://nuxt.com "mini") suite'), {
    type: 'linkCard',
    raw: '[Nuxt](https://nuxt.com "mini")',
    text: 'Nuxt',
    href: 'https://nuxt.com',
    size: 'mini',
  });
  assert.equal(readLinkCard('[Nuxt](http://nuxt.com/a?b=1 "Large")')?.size, 'large');
  assert.equal(readLinkCard('voir [Nuxt](https://nuxt.com "mini")'), undefined);
});

test('leaves hyperlinks and other links to markdown', () => {
  for (const src of [
    '[Nuxt](https://nuxt.com)',
    '[Nuxt](https://nuxt.com "autre")',
    '[x](javascript:alert(1) "mini")',
    '[x](note:3f2a0000-0000-0000-0000-00000000abcd "mini")',
    '[x](mailto:a@b.c "large")',
  ]) {
    assert.equal(readLinkCard(src), undefined, src);
  }
});

test('writes a card that reads back the same', () => {
  const md = writeLinkCard('Doc [v4]', 'https://ui.nuxt.com/a b(c)"', 'large');
  assert.equal(md, '[Doc v4](https://ui.nuxt.com/a%20b%28c%29%22 "large")');
  assert.deepEqual(readLinkCard(md), {
    type: 'linkCard',
    raw: md,
    text: 'Doc v4',
    href: 'https://ui.nuxt.com/a%20b%28c%29%22',
    size: 'large',
  });
});

test('only http and https become cards', () => {
  assert.equal(isWebLink('https://a.b'), true);
  assert.equal(isWebLink('mailto:a@b.c'), false);
  assert.equal(isWebLink(undefined), false);
});

test('a link in a size: a card for the web, a hyperlink otherwise', () => {
  assert.deepEqual(linkContent('https://ui.nuxt.com/x', ''), {
    type: 'linkCard',
    attrs: { href: 'https://ui.nuxt.com/x', text: 'ui.nuxt.com', size: DEFAULT_LINK_SIZE },
  });
  assert.equal(linkContent('https://a.b', 'A', 'large').attrs?.size, 'large');
  assert.deepEqual(linkContent('https://a.b', 'A', 'hyperlink'), {
    type: 'text',
    text: 'A',
    marks: [{ type: 'link', attrs: { href: 'https://a.b' } }],
  });
  assert.equal(linkContent('mailto:a@b.c', '', 'large').type, 'text');
});
