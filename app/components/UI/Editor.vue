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
</style>
