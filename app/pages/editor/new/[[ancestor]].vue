<script setup lang="ts">
import type { EditorProps } from '~/composables/editor/useEditor';

/* Define */
const { $hooks } = useNuxtApp();
const route = useRoute();
const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

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
  if (!content.value) return;

  if (folder.value) {
    await $hooks.callHook('data:create:note:in-folder', {
      folder: folder.value,
      content: content.value,
    });
    return;
  }

  const result = await $hooks.callHook('data:create:note', {
    type: ItemType.NOTE,
    content: content.value,
    path: [],
  });
  logSubmit.info('created', result?.id);
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
