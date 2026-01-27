<script setup lang="ts">
// Define
import type { EditorProps } from '~/components/Editor.vue';

const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

const { handleCreate } = useDataApi();

const content = ref<string | null>(null);

const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
// Data

// Methods
const handleSubmit = async () => {
  if (!content.value) return;

  await handleCreate({
    path: [], // TODO set path autrement
    type: ItemType.NOTE,
    content: content.value,
  });
};

// Lifecycle

// SEO
</script>

<template>
  <div>
    <p>FORM NEW</p>
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
