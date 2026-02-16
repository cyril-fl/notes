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

const ancestorId = computed<string | null>(() => {
  const param = route.params.ancestor;
  if (!param || Array.isArray(param)) return null;
  return param;
});

const folder = computed<Folder | undefined>(() => {
  if (!ancestorId.value) return undefined;
  return getById(ancestorId.value, { types: ItemType.FOLDER });
});

const content = ref<string | undefined>();
const editorKey = ref<string>(
  ancestorId.value ? `editor-new-${ancestorId.value}` : 'editor-new-root'
);
const isSwitchingContext = ref(false);

watch(
  () => ancestorId.value,
  (next, prev) => {
    if (next === prev) return;
    isSwitchingContext.value = true;
    content.value = undefined;
    editorKey.value = next
      ? `editor-new-${next}`
      : `editor-new-root-${Date.now()}`;
  }
);

watch(
  () => folder.value,
  () => {
    if (isSwitchingContext.value) {
      isSwitchingContext.value = false;
    }
  },
  { immediate: true }
);

// const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
/* Data */

/* Methods */
const handleSubmit = async () => {
  if (isSwitchingContext.value) return;
  if (!content.value) return;

  if (folder.value) {
    await $hooks.callHook('data:create:note:in-folder', {
      folder: folder.value,
      content: content.value,
    });
    return;
  }

  const result = await $hooks.callHook('data:create:note', {
    content: content.value,
    path: [],
  });

  logSubmit.info('created', result?.id);
};

/* Lifecycle */

/* SEO */
</script>

<template>
  <UIPageSection :title="$t('pages.editor.new.title')">
    <UIEditor
      :key="editorKey"
      v-model:content="content"
      v-bind="props"
      @submit="handleSubmit"
      @update:mentions="onUpdateMentions"
    />
  </UIPageSection>
</template>

<style scoped></style>
