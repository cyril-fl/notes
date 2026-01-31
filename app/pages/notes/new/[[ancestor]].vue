<script setup lang="ts">
// Define
import type { EditorProps } from '~/components/Editor.vue';

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

const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
// Data

// Methods
const handleSubmit = async () => {
  if (!content.value) return;

  if (folder.value) {
    await handleCreateNoteInFolder({
      folder: folder.value,
      content: content.value,
    });
    return;
  }

  await handleCreate({
    type: ItemType.NOTE,
    content: content.value,
    path: [],
  });
};

// Lifecycle

// SEO
</script>

<template>
  <div>
    <p>FORM NEW {{ route.params.ancestor }}</p>
    <Editor
      v-model:content="content"
      v-bind="props"
      @submit="handleSubmit"
      @update:hashtags="onUpdateHashtags"
      @update:mentions="onUpdateMentions"
    />
  </div>
</template>

<style scoped></style>
