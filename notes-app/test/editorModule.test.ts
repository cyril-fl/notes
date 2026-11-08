import assert from 'node:assert/strict';
import { test } from 'node:test';
import { command, mergeModules } from '../app/utils/editorModule.ts';

test('a command can always run and is inactive by default', () => {
  const handler = command(() => 'ran');
  assert.equal(handler.canExecute(null as never), true);
  assert.equal(handler.isActive(null as never), false);
  assert.equal(handler.execute(null as never), 'ran');
  assert.equal(
    command(
      () => 0,
      () => true,
    ).isActive(null as never),
    true,
  );
});

test('merges the modules in order', () => {
  const clicks: string[] = [];
  const merged = mergeModules([
    {
      extensions: ['a' as never],
      starterKit: { underline: false },
      handlers: { one: command(() => 1) },
      toolbar: [[{ kind: 'bold' } as never]],
      onClick: () => clicks.push('first'),
    },
    {},
    {
      extensions: ['b' as never],
      starterKit: { link: {} },
      handlers: { two: command(() => 2) },
      toolbar: [[{ kind: 'link' } as never]],
      menus: [{ char: '#', pluginKey: 'tagMenu', items: [] }],
      onClick: () => clicks.push('last'),
    },
  ]);
  assert.deepEqual(merged.extensions, ['a', 'b']);
  assert.deepEqual(merged.starterKit, { underline: false, link: {} });
  assert.deepEqual(Object.keys(merged.handlers), ['one', 'two']);
  assert.equal(merged.toolbar.length, 2);
  assert.equal(merged.menus.length, 1);
  assert.deepEqual(merged.bubbles, []);
  merged.onClick(new Event('click') as MouseEvent);
  assert.deepEqual(clicks, ['first', 'last']);
});
