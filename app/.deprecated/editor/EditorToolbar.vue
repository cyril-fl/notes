<script lang="ts" setup>
import type { Editor } from '@tiptap/core';
import type { BaseEditorToolbarItem } from '~/types/ui';

const props = defineProps<{
  editor: Editor | null;
  items: BaseEditorToolbarItem[][];
}>();

function runCommand(item: BaseEditorToolbarItem) {
  const editor = props.editor;
  if (!editor) return;

  const chain = editor.chain().focus();

  switch (item.kind) {
    case 'undo':
      editor.commands.undo();
      break;
    case 'redo':
      editor.commands.redo();
      break;
    case 'mark':
      if (item.mark === 'bold') chain.toggleBold().run();
      else if (item.mark === 'italic') chain.toggleItalic().run();
      else if (item.mark === 'underline') chain.toggleUnderline().run();
      else if (item.mark === 'strike') chain.toggleStrike().run();
      else if (item.mark === 'code') chain.toggleCode().run();
      break;
    case 'heading':
      chain.toggleHeading({ level: (item.level ?? 1) as 1 | 2 | 3 }).run();
      break;
    case 'textAlign':
      chain.setTextAlign(item.align ?? 'left').run();
      break;
    case 'bulletList':
      chain.toggleBulletList().run();
      break;
    case 'orderedList':
      chain.toggleOrderedList().run();
      break;
    case 'blockquote':
      chain.toggleBlockquote().run();
      break;
    case 'link':
      chain.setLink({ href: '' }).run();
      break;
    default:
      break;
  }
}
</script>

<template>
  <div
    v-if="editor"
    class="flex flex-wrap gap-1 border-b border-default bg-default p-1"
  >
    <template v-for="(group, gi) in items" :key="gi">
      <template v-for="(item, i) in group" :key="`${gi}-${i}`">
        <button
          :title="item.tooltip?.text"
          class="rounded p-1.5 text-muted hover:bg-muted hover:text-default"
          type="button"
          @click="runCommand(item)"
        >
          <Icon v-if="item.icon" :name="item.icon" class="size-4" />
        </button>
      </template>
      <span
        v-if="gi < items.length - 1"
        class="mx-0.5 w-px self-stretch border-l border-default"
      />
    </template>
  </div>
</template>
