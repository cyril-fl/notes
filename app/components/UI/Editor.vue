<script lang="ts" setup>
import type { EditorProps } from '~/composables/editor/useEditor';
import {  useProvideEditorContext} from '~/composables/editor/useEditor';
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

const { content, editor, preview } = useProvideEditorContext(props);

const model = defineModel<string | null>({
  default: null,
});
/* Data */

/* Methods */
const handleSubmit = () => {
  emit('submit');
};

/* Lifecycle */
watch(
  () => content.value,
  () => {
    model.value = content.value;
  }
);

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
  <div
    class="grow bg-plain-100/20 text-plain-500 rounded-2xl p-4 overflow-hidden"
  >
    <UIEditorToolBar />
    <EditorContent
      v-model="model"
      :editor="editor"
      class="input size-full"
    />
  </div>
  <ul class="mt-4">
      <template v-for="item in preview" :key="item.label">
        <li v-if="item.condition">
          <p>{{ item.label }} :</p>
          <pre>{{ item.content }}</pre>
        </li>
      </template>
    </ul>
</template>

<style lang="css" scoped>
/* Retirer le contour Safari par défaut sur les champs actifs / contenteditable */
input:focus,
input:focus-visible {
  outline: none;
  box-shadow: none;
}
/* ProseMirror (TipTap) : retirer le contour bleu Safari sur l’éditeur */
:deep(.ProseMirror),
:deep(.ProseMirror:focus),
:deep(.ProseMirror:focus-visible),
:deep([contenteditable='true']:focus) {
  outline: none !important;
  box-shadow: none !important;
}

:deep(.ProseMirror > h1) {
  font-size: 1.5rem;
  line-height: 2rem;
  font-weight: 700;
}
:deep(.ProseMirror > h2) {
  font-size: 1.25rem;
  line-height: 1.75rem;
  font-weight: 700;
}
:deep(.ProseMirror > h3) {
  font-size: 1.125rem;
  line-height: 1.75rem;
  font-weight: 700;
}
:deep(.ProseMirror > h4) {
  font-size: 1rem;
  line-height: 1.5rem;
  font-weight: 700;
}
:deep(.ProseMirror > h5) {
  font-size: 0.875rem;
  line-height: 1.25rem;
  font-weight: 700;
}
:deep(.ProseMirror > h6) {
  font-size: 0.75rem;
  line-height: 1rem;
  font-weight: 700;
}
:deep(.ProseMirror > p) {
  font-size: 1rem;
  line-height: 1.5rem;
  font-weight: 400;
}
:deep(.ProseMirror > ul) {
  list-style-type: disc;
}
:deep(.ProseMirror > ol) {
  list-style-type: decimal;
}
:deep(.ProseMirror > li) {
  margin-left: 1rem;
}
:deep(.ProseMirror > blockquote) {
  margin-left: 1rem;
  border-left: 2px solid #ccc;
  padding-left: 1rem;
}
:deep(.ProseMirror > code) {
  font-size: 0.875rem;
  line-height: 1.25rem;
  font-weight: 400;
}
</style>
