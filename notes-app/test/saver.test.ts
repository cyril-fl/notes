import assert from 'node:assert/strict';
import { test } from 'node:test';
import { createSaver, type Changes } from '../app/utils/saver.ts';

// A note that POSTs first then PATCHes, each request waiting until release() is called.
function fakeNote() {
  const calls: { method: string; changes: Changes }[] = [];
  const releases: (() => void)[] = [];
  let id: string | null = null;

  const send = async (changes: Changes) => {
    calls.push({ method: id ? 'PATCH' : 'POST', changes });
    await new Promise<void>((resolve) => releases.push(resolve));
    id = 'created';
  };
  const release = async () => {
    releases.shift()?.();
    await new Promise((resolve) => setImmediate(resolve));
  };
  return { calls, send, release };
}

test('edits during a POST give one POST then one PATCH', async () => {
  const { calls, send, release } = fakeNote();
  const saver = createSaver(send, 1000);

  saver.push({ content: 'a' });
  const done = saver.flush();

  saver.push({ content: 'ab' });
  saver.push({ title: 'T', content: 'abc' });
  saver.flush();

  await release();
  await release();
  await done;

  assert.deepEqual(calls, [
    { method: 'POST', changes: { content: 'a' } },
    { method: 'PATCH', changes: { title: 'T', content: 'abc' } },
  ]);
});

test('flush sends what waits without the delay', async () => {
  const sent: Changes[] = [];
  const saver = createSaver(async (changes) => void sent.push(changes), 60_000);

  saver.push({ content: 'a' });
  await saver.flush();

  assert.deepEqual(sent, [{ content: 'a' }]);
});

test('a failure keeps the changes, newer ones win', async () => {
  const sent: Changes[] = [];
  let fail = true;
  const saver = createSaver(async (changes) => {
    if (fail) throw new Error('down');
    sent.push(changes);
  }, 60_000);

  saver.push({ title: 'old', content: 'a' });
  await saver.flush();
  assert.deepEqual(sent, []);

  fail = false;
  saver.push({ content: 'ab' });
  await saver.flush();

  assert.deepEqual(sent, [{ title: 'old', content: 'ab' }]);
});

test('send returning false keeps the changes without retrying in a loop', async () => {
  let calls = 0;
  const saver = createSaver(async () => {
    calls++;
    return false;
  }, 60_000);

  saver.push({ title: 'T' });
  await saver.flush();

  assert.equal(calls, 1);
  assert.deepEqual(saver.take(), { title: 'T' });
});

test('take empties the queue', async () => {
  let calls = 0;
  const saver = createSaver(async () => void calls++, 60_000);

  saver.push({ content: 'a' });
  assert.deepEqual(saver.take(), { content: 'a' });
  await saver.flush();

  assert.equal(calls, 0);
});
