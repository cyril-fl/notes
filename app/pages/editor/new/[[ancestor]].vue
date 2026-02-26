<script setup lang="ts">
import type { EditorProps } from '~/composables/editor/useEditor';

const route = useRoute();
const props = defineProps<Partial<EditorProps>>();

const { getById } = useDataUtils();
const { createNoteInFolder, createNote } = useDataActions();

const ancestorId = computed<string | null>(() => {
  const param = route.params.ancestor;
  if (!param || Array.isArray(param)) return null;
  return param;
});

const folder = computed<Folder | undefined>(() => {
  if (!ancestorId.value) return undefined;
  return getById(ancestorId.value, { types: ItemType.FOLDER });
});

usePageSection(
  computed(() => ({
    searchable: false,
    ancestors: folder.value?.path,
  }))
);

const content = ref<string | undefined>();
const editorKey = ref<string>(
  ancestorId.value ? `editor-new-${ancestorId.value}` : 'editor-new-root'
);
const isSwitchingContext = ref(false);

const { onUpdate: onUpdateMentions } = useMentions();

const handleSubmit = async () => {
  if (isSwitchingContext.value) return;
  if (!content.value) return;

  if (folder.value) {
    await createNoteInFolder({
      folder: folder.value,
      content: content.value,
    });
    return;
  }

  const result = await createNote({
    content: content.value,
    path: [],
  });

  logSubmit.info('created', result?.id);
};

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
</script>

<template>
  <UIEditor
    :key="editorKey"
    v-model:content="content"
    v-bind="props"
    @submit="handleSubmit"
    @update:mentions="onUpdateMentions"
  />
</template>

<style scoped></style>
