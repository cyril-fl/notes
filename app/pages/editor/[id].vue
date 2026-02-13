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

const { handleUpdate } = useDataApi();
const { getById } = useDataUtils();

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

const _content = ref<string | null>(note.value?.content ?? null);
const content = computed({
  get: () => {
    return _content.value;
  },
  set: (value: string | null) => {
    _content.value = value;
  },
});

let hasInitializedContent = !!_content.value;

watch(
  () => note.value?.content,
  (noteContent) => {
    if (hasInitializedContent || noteContent === undefined) return;
    hasInitializedContent = true;
    _content.value = noteContent ?? null;
  },
  { immediate: true }
);

// const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
/* Data */

/* Methods */
// const handleSubmit = async () => {
//   if (!(note.value?.id && content.value)) return;
//   await handleUpdate(note.value.id, {
//     content: content.value,
//   });
// };

/* Lifecycle */
watch(
  () => ({ note: note.value, hasLoaded: hasLoaded.value }),
  ({ note, hasLoaded }) => {
    if (note || !hasLoaded) return;
    navigateTo('/notes/new', { replace: true });
  },
  { immediate: true }
);

/* SEO */
</script>

<template>
  <section class="grow flex flex-col">
    <UIEditor
      v-bind="props"
      v-model:content="content"
      @update:mentions="onUpdateMentions"
    />
    <!-- v-model:content="content" -->
    <!-- @submit="handleSubmit" -->
  </section>
</template>
