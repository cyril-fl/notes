<script lang="ts" setup>
import type { EditorProps } from '~/composables/editor/useEditor';
import { useProvideEditorContext } from '~/composables/editor/useEditor';
import { EditorContent } from '@tiptap/vue-3';

interface Emits {
  (e: 'submit'): void;
  // (e: 'update:hashtags' | 'update:mentions', value: string[]): void;
}

const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: false,
  showMentions: false,
  showPreview: false,
});

const emit = defineEmits<Emits>();

const model = defineModel<string | undefined>('content');

const { editor, preview } = useProvideEditorContext(props, model);

const handleSubmit = () => {
  emit('submit');
};

onBeforeRouteLeave(handleSubmit);

onMounted(() => window.addEventListener('beforeunload', handleSubmit));

onBeforeUnmount(() => window.removeEventListener('beforeunload', handleSubmit));
</script>

<template>
  <div
    class="editor-content relative bg-muted grow rounded-2xl p-4 overflow-x-hidden overflow-y-auto"
  >
    <UICoreToolbar>
      <template #right>
        <UIEditorToolBar />
      </template>
    </UICoreToolbar>

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
