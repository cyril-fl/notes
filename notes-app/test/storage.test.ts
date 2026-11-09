import assert from 'node:assert/strict';
import { test } from 'node:test';
import { fileKind, fileSize } from '../app/utils/storage.ts';

test('a file is sorted by its content type, a zip and an Ogg by what they hold most', () => {
  assert.deepEqual(
    ['image/png', 'application/pdf', 'application/zip', 'audio/mpeg', 'application/ogg', 'video/webm'].map(fileKind),
    ['images', 'documents', 'documents', 'audio', 'audio', 'video'],
  );
});

test('a size is written in the largest unit under 1000', () => {
  assert.deepEqual(
    [fileSize(0, 'en'), fileSize(999, 'en'), fileSize(1500, 'en'), fileSize(20_000_000, 'en'), fileSize(3e12, 'en')],
    ['0 byte', '999 byte', '1.5 kB', '20 MB', '3,000 GB'],
  );
});
