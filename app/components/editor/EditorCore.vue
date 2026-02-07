<script lang="ts" setup>
import { EditorContent, useEditor } from '@tiptap/vue-3';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import Link from '@tiptap/extension-link';
import Markdown from '@tiptap/markdown';
import Emoji from '@tiptap/extension-emoji';
import TextAlign from '@tiptap/extension-text-align';
import Highlight from '@tiptap/extension-highlight';

const props = withDefaults(
  defineProps<{
    content: string | null;
    placeholder?: string;
    readonly?: boolean;
  }>(),
  { placeholder: '', readonly: false }
);

const emit = defineEmits<{ 'update:content': [value: string | null] }>();

const emojiItems = useEditorEmoji();
const mentionsItems = useEditorMention();
const toolBarItems = useEditorToolBar();

const extensions = [
  StarterKit,
  Markdown,
  Placeholder.configure({ placeholder: props.placeholder }),
  Link.configure({ openOnClick: false }),
  Emoji,
  TextAlign,
  Highlight,
];

const lastEmittedContent = ref<string | null>(null);

const editor = useEditor({
  content: props.content ?? undefined,
  contentType: 'markdown',
  editable: !props.readonly,
  extensions,
  editorProps: {
    attributes: {
      class:
        'prose prose-sm max-w-none min-h-[5rem] w-full px-3 py-2 outline-none',
    },
  },
  onUpdate: ({ editor: e }) => {
    const md = (e as unknown as { getMarkdown?: () => string }).getMarkdown?.();
    lastEmittedContent.value = md ?? null;
    emit('update:content', lastEmittedContent.value);
  },
});

watch(
  () => props.content,
  (newContent) => {
    if (!editor.value || newContent === undefined || newContent === null)
      return;
    if (lastEmittedContent.value === newContent) return;
    const current = (
      editor.value as unknown as { getMarkdown?: () => string }
    ).getMarkdown?.();
    if (current !== newContent) {
      lastEmittedContent.value = newContent;
      editor.value.commands.setContent(newContent, {
        contentType: 'markdown',
        emitUpdate: false,
      });
    }
  },
  { flush: 'post' }
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});
</script>

<template>
  <div
    class="bg-muted w-full min-h-21 overflow-hidden rounded-md border border-default"
  >
    <template v-if="editor">
      <EditorToolbar v-if="!readonly" :editor="editor" :items="toolBarItems" />
      <EditorEmojiMenu :editor="editor" :items="emojiItems" />
      <EditorMentionMenu :editor="editor" :items="mentionsItems" />
    </template>
    <EditorContent :editor="editor" />
  </div>
</template>
