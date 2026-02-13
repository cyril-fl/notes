<script setup lang="ts">
import type { EditorProps } from '~/composables/editor/useEditor';

/* Define */
const route = useRoute();
const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

const { handleCreate, handleCreateNoteInFolder } = useDataApi();
const { getById } = useDataUtils();

const folder = computed<Folder | undefined>(() => {
  const ancestorParam = route.params.ancestor;
  if (!ancestorParam || typeof ancestorParam !== 'string') return undefined;
  return getById(ancestorParam, { types: ItemType.FOLDER });
});

const content = ref<string | null>(null);

// const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
/* Data */

/* Methods */
const handleSubmit = async () => {
  console.log('[HANDLE SUBMIT] - content', content.value);
  if (!content.value) return;

  if (folder.value) {
    await handleCreateNoteInFolder({
      folder: folder.value,
      content: content.value,
    });
    return;
  }

  const result = await handleCreate({
    type: ItemType.NOTE,
    content: content.value,
    path: [],
  });
  console.log('[HANDLE SUBMIT] - result', result);
};

/* Lifecycle */

/* SEO */
</script>

<template>
  <section class="grow flex flex-col">
    <UIEditor
      v-model:content="content"
      v-bind="props"
      @submit="handleSubmit"
      @update:mentions="onUpdateMentions"
    />
  </section>
</template>

<style scoped></style>
