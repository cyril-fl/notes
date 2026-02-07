<script lang="ts" setup>
import type { Editor } from '@tiptap/core';
import type { EditorMentionMenuItem } from '~/types/ui';

const props = defineProps<{
  editor: Editor | null;
  items: EditorMentionMenuItem[];
  appendTo?: () => HTMLElement;
}>();

function insertMention(item: EditorMentionMenuItem) {
  const editor = props.editor;
  if (!editor) return;
  editor.chain().focus().insertContent(`@${item.label} `).run();
}
</script>

<template>
  <div
    v-if="editor && items.length"
    class="flex flex-wrap gap-1 border-b border-default bg-default p-1"
  >
    <button
      v-for="item in items"
      :key="item.id"
      class="rounded px-2 py-1 text-sm text-muted hover:bg-muted hover:text-default"
      type="button"
      @click="insertMention(item)"
    >
      {{ item.label }}
    </button>
  </div>
</template>
