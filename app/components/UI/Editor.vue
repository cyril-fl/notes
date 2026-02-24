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
    <div
      class="flex items-center gap-2 sticky top-0 z-10 py-2 bg-muted/80 backdrop-blur-sm"
    >
      <UIPageTitle
        :title="props.title ?? $t('notes.default_title')"
        class="grow min-w-0"
      />
      <UIEditorToolBar />
    </div>
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
