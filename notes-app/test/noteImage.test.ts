import assert from 'node:assert/strict';
import { test } from 'node:test';
import { citedSrc, imageSize, imageSrc } from '../app/utils/noteImage.ts';

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
  for (const title of [null, undefined, '', 'large', 'Mon chat']) assert.equal(imageSize(title), null, String(title));
});
