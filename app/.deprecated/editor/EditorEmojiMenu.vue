<script lang="ts" setup>
import type { Editor } from '@tiptap/core';
import type { EditorEmojiMenuItem } from '~/types/ui';

const props = defineProps<{
  editor: Editor | null;
  items: EditorEmojiMenuItem[];
  appendTo?: () => HTMLElement;
}>();

function insertEmoji(item: EditorEmojiMenuItem) {
  const editor = props.editor;
  if (!editor) return;
  const value = item.emoji ?? item.name;
  editor.chain().focus().insertContent(value).run();
}
</script>

<template>
  <div
    v-if="editor && items.length"
    class="flex flex-wrap gap-0.5 border-b border-default bg-default p-1"
  >
    <button
      v-for="(item, i) in items.slice(0, 24)"
      :key="item.name"
      :title="item.name"
      class="rounded p-1 text-lg hover:bg-muted"
      type="button"
      @click="insertEmoji(item)"
    >
      {{ item.emoji ?? item.name }}
    </button>
  </div>
</template>
