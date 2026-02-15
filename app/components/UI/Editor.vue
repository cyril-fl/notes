<script lang="ts" setup>
import type { EditorProps } from '~/composables/editor/useEditor';
import { useProvideEditorContext } from '~/composables/editor/useEditor';
import { EditorContent } from '@tiptap/vue-3';
/* Define */

interface Emits {
  (e: 'submit'): void;
  // (e: 'update:hashtags' | 'update:mentions', value: string[]): void;
}

const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

const emit = defineEmits<Emits>();

const model = defineModel<string | undefined>('content');

const { editor, preview } = useProvideEditorContext(props, model);

/* Data */

/* Methods */
const handleSubmit = () => {
  emit('submit');
};

onBeforeRouteLeave(handleSubmit);

onMounted(() => {
  window.addEventListener('beforeunload', handleSubmit);
});

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleSubmit);
});
/* SEO */
</script>

<template>
  <div class="editor-content bg-muted grow rounded-2xl p-4 overflow-hidden">
    <UIEditorToolBar />
    <EditorContent :editor="editor" class="size-full" />
  </div>
  <ul v-if="preview.some((item) => item.condition)" class="mt-4">
    <template v-for="item in preview" :key="item.label">
      <li v-if="item.condition">
        <p>{{ item.label }} :</p>
        <pre>{{ item.content }}</pre>
      </li>
    </template>
  </ul>
</template>

<style lang="css" scoped>
@reference "../../assets/css/main.css";

/* RACINE */
:deep(.ProseMirror) {
  @apply min-h-64 outline-none px-1 py-2;
}

:deep(.ProseMirror:focus),
:deep(.ProseMirror:focus-visible),
:deep([contenteditable='true']:focus) {
  @apply outline-none shadow-none;
}

/* PLACEHOLDER */
:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  @apply text-dimmed float-left h-0 pointer-events-none;
  content: attr(data-placeholder);
}

/* ESPACEMENT VERTICAL ENTRE BLOCS */
:deep(.ProseMirror > * + *) {
  @apply mt-2;
}

:deep(.ProseMirror > * + h1),
:deep(.ProseMirror > * + h2),
:deep(.ProseMirror > * + h3) {
  @apply mt-6;
}

:deep(.ProseMirror > * + h4),
:deep(.ProseMirror > * + h5),
:deep(.ProseMirror > * + h6) {
  @apply mt-4;
}

/* TITRES */
:deep(.ProseMirror h1) {
  @apply text-2xl font-bold leading-tight tracking-tight;
}

:deep(.ProseMirror h2) {
  @apply text-xl font-bold leading-snug;
}

:deep(.ProseMirror h3) {
  @apply text-lg font-semibold leading-snug;
}

:deep(.ProseMirror h4) {
  @apply text-base font-semibold leading-normal;
}

:deep(.ProseMirror h5) {
  @apply text-sm font-semibold leading-normal uppercase tracking-wide;
}

:deep(.ProseMirror h6) {
  @apply text-xs font-semibold leading-normal uppercase tracking-wider text-muted;
}

/* PARAGRAPHE */
:deep(.ProseMirror p) {
  @apply text-base leading-relaxed;
}

/* MARKS INLINE */
:deep(.ProseMirror strong) {
  @apply font-bold;
}

:deep(.ProseMirror em) {
  @apply italic;
}

:deep(.ProseMirror u) {
  @apply underline underline-offset-2;
}

:deep(.ProseMirror del),
:deep(.ProseMirror s) {
  @apply line-through text-muted;
}

/* CODE INLINE */
:deep(.ProseMirror code) {
  @apply font-mono text-sm bg-accented px-1.5 py-0.5 rounded-md;
}

/* CODE BLOCK */
:deep(.ProseMirror pre) {
  @apply bg-accented rounded-lg p-4 overflow-x-auto my-3;
}

:deep(.ProseMirror pre code) {
  @apply bg-transparent p-0 rounded-none text-sm leading-relaxed;
}

/* LIENS */
:deep(.ProseMirror a) {
  @apply text-primary-500 no-underline cursor-pointer;
}

:deep(.ProseMirror a:hover) {
  @apply underline underline-offset-2;
}

:deep(.ProseMirror a:focus-visible) {
  @apply outline-2 outline-offset-2 outline-primary-500 rounded-sm;
}

/* LISTES (bullet + ordered) */
:deep(.ProseMirror ul) {
  @apply list-disc pl-6;
}

:deep(.ProseMirror ol) {
  @apply list-decimal pl-6;
}

:deep(.ProseMirror li) {
  @apply leading-relaxed;
}

:deep(.ProseMirror li p) {
  @apply my-0;
}

:deep(.ProseMirror li > ul),
:deep(.ProseMirror li > ol) {
  @apply mt-1;
}

/* TASK LIST */
:deep(.ProseMirror ul[data-type='taskList']) {
  @apply list-none pl-0;
}

:deep(.ProseMirror ul[data-type='taskList'] li) {
  @apply flex items-start gap-2;
}

:deep(.ProseMirror ul[data-type='taskList'] li > label) {
  @apply shrink-0 mt-1;
}

:deep(.ProseMirror ul[data-type='taskList'] li > label input[type='checkbox']) {
  @apply size-4 rounded cursor-pointer accent-primary-500;
  border: 2px solid var(--ui-border-accented);
}

:deep(.ProseMirror ul[data-type='taskList'] li > div) {
  @apply flex-1 min-w-0;
}

:deep(.ProseMirror ul[data-type='taskList'] li[data-checked='true'] > div) {
  @apply line-through text-dimmed;
}

:deep(.ProseMirror ul[data-type='taskList'] ul[data-type='taskList']) {
  @apply pl-6 mt-1;
}

/* BLOCKQUOTE */
:deep(.ProseMirror blockquote) {
  @apply pl-4 italic my-3;
  border-left: 3px solid var(--ui-border-accented);
}

:deep(.ProseMirror blockquote p) {
  @apply text-muted;
}

/* HORIZONTAL RULE */
:deep(.ProseMirror hr) {
  @apply my-6 border-none;
  border-top: 1px solid var(--ui-border);
}

/* HIGHLIGHT */
:deep(.ProseMirror mark) {
  @apply rounded-sm px-0.5;
  background-color: var(--color-warning-200);
}

/* SELECTION */
:deep(.ProseMirror ::selection) {
  @apply bg-primary-100;
}

/* GAPCURSOR */
:deep(.ProseMirror .ProseMirror-gapcursor::after) {
  border-top-color: var(--ui-text);
}
</style>
