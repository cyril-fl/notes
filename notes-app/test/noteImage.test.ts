import assert from 'node:assert/strict';
import { test } from 'node:test';
import { citedSrc, imageSize, imageSrc, insertFiles, openFile } from '../app/utils/noteImage.ts';

const API = 'http://localhost:8080';
const FILE = '/api/files/3f2a0000-0000-0000-0000-00000000abcd';

test('shows the files of the API from the API, other images as written', () => {
  assert.equal(imageSrc(FILE, API), API + FILE);
  assert.equal(imageSrc('https://x.org/a.png', API), 'https://x.org/a.png');
});

test('cites the files of the API by their path, whatever the address they come with', () => {
  assert.equal(citedSrc(API + FILE, API), FILE);
  assert.equal(citedSrc(FILE, API), FILE);
  assert.equal(citedSrc('https://x.org/api/files/a.png', API), 'https://x.org/api/files/a.png');
});

test('reads a size from the title of an image, any other title is not one', () => {
  assert.equal(imageSize('small'), 'small');
  assert.equal(imageSize('Medium'), 'medium');
  assert.equal(imageSize('large'), 'large');
  for (const title of [null, undefined, '', 'huge', 'Mon chat']) assert.equal(imageSize(title), null, String(title));
});

test('opens a file of a note in a new tab, so an error of the API does not replace the note', () => {
  const opened: unknown[][] = [];
  globalThis.window = { open: (...args: unknown[]) => opened.push(args) } as never;
  openFile(FILE, API);
  assert.deepEqual(opened, [[API + FILE, '_blank', 'noopener']]);
});

/** An editor writing down what it is asked to insert, destroyed as soon as the first file is sent. */
function stubEditor(inserted: unknown[], destroy?: () => boolean) {
  const chain = { focus: () => chain, insertContent: (c: unknown) => (inserted.push(c), chain), run: () => true };
  return {
    get isDestroyed() {
      return destroy?.() ?? false;
    },
    chain: () => chain,
  } as never;
}

test('writes the sent files in the note, and nothing once the note is left', async () => {
  const file = new File(['x'], 'cat.png', { type: 'image/png' });

  const inserted: unknown[] = [];
  await insertFiles(stubEditor(inserted), async () => FILE, [file]);
  assert.deepEqual(inserted, [{ type: 'image', attrs: { src: FILE, alt: 'cat.png' } }]);

  const none: unknown[] = [];
  let left = false;
  await insertFiles(
    stubEditor(none, () => left),
    async () => ((left = true), FILE),
    [file, file],
  );
  assert.deepEqual(none, []);
});
