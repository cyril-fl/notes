import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  DEFAULT_LINK_SIZE,
  isCardLink,
  linkContent,
  linkHost,
  readLinkCard,
  toHref,
  writeLinkCard,
} from '../app/utils/linkCard.ts';

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

test('web and mail links become cards', () => {
  assert.equal(isCardLink('https://a.b'), true);
  assert.equal(isCardLink('MAILTO:a@b.c'), true);
  assert.equal(isCardLink('tel:0600'), false);
  assert.equal(isCardLink(undefined), false);
  assert.equal(readLinkCard('[Moi](mailto:a@b.c "large")')?.href, 'mailto:a@b.c');
});

test('a mail link shows its address, without the subject', () => {
  assert.equal(linkHost('mailto:jean%2Bnotes@b.c?subject=Salut'), 'jean+notes@b.c');
  assert.equal(linkHost('https://ui.nuxt.com/x'), 'ui.nuxt.com');
  assert.equal(linkHost('mailto:a%ZZ@b.c'), 'mailto:a%ZZ@b.c');
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
  assert.equal(linkContent('mailto:a@b.c', '', 'large').attrs?.text, 'a@b.c');
  assert.equal(linkContent('tel:0600', '', 'large').type, 'text');
});

test('an address typed for the link button gets its scheme', () => {
  assert.equal(toHref(' nasfwg@gmail.com '), 'mailto:nasfwg@gmail.com');
  assert.equal(toHref('example.com/a'), 'https://example.com/a');
  assert.equal(toHref('https://user@example.com'), 'https://user@example.com');
  assert.equal(toHref('mailto:a@b.c'), 'mailto:a@b.c');
});
