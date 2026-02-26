<script setup lang="ts">
import type { EditorProps } from '~/composables/editor/useEditor';

const route = useRoute();
const props = defineProps<Partial<EditorProps>>();

usePageSection({ searchable: false });

const { getById } = useDataUtils();
const { update } = useDataActions();

const store = useDataStore();
const { hasLoaded } = storeToRefs(store);

const note = computed(() => {
  const id = route.params.id;
  const isIdValid = id && typeof id === 'string';

  return isIdValid
    ? getById(id as string, {
        types: ItemType.NOTE,
      })
    : null;
});

const _content = ref<string | undefined>(note.value?.content);
const content = computed({
  get: () => {
    return _content.value;
  },
  set: (value: string | undefined) => {
    _content.value = value;
  },
});

let hasInitializedContent = !!_content.value;

const { onUpdate: onUpdateMentions } = useMentions();

const handleSubmit = async () => {
  if (!(note.value?.id && content.value)) return;
  await update(note.value.id, { content: content.value });
};

watch(
  () => note.value?.content,
  (noteContent) => {
    if (hasInitializedContent) return;
    hasInitializedContent = true;
    _content.value = noteContent;
  },
  { immediate: true }
);

watch(
  () => ({ note: note.value, hasLoaded: hasLoaded.value }),
  ({ note, hasLoaded }) => {
    if (note || !hasLoaded) return;
    navigateTo(NAVIGATION.newNote, { replace: true });
  },
  { immediate: true }
);
</script>

<template>
  <UIEditor
    v-bind="props"
    v-model:content="content"
    @update:mentions="onUpdateMentions"
    @submit="handleSubmit"
  />
</template>
