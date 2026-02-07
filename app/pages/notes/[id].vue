<script setup lang="ts">
// Define
import type { EditorProps } from '~/components/Editor.vue';

const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

const route = useRoute();
const store = useDataStore();
const { hasLoaded } = storeToRefs(store);

const { getById } = useDataUtils();

const note = computed(() => {
  const id = route.params.id;
  const isIdValid = id && typeof id === 'string';

  return isIdValid
    ? getById(id as string, {
        types: ItemType.NOTE,
      })
    : null;
});

const { handleUpdate } = useDataApi();

const content = ref<string | null>(note.value?.content ?? null);

// const { onUpdate: onUpdateHashtags } = useHashtags();
const { onUpdate: onUpdateMentions } = useMentions();
// Data

// Methods
const handleSubmit = async () => {
  if (!(note.value?.id && content.value)) return;
  await handleUpdate(note.value.id, {
    content: content.value,
  });
};

// Lifecycle
watch(
  () => ({ note: note.value, hasLoaded: hasLoaded.value }),
  ({ note: n, hasLoaded: ok }) => {
    if (!ok || n) return;
    navigateTo('/notes/new', { replace: true });
  },
  { immediate: true }
);

// SEO
</script>

<template>
  <section>
    <p>FORM EDIT</p>
    <Editor
      v-if="note"
      v-model:content="content"
      v-bind="props"
      @submit="handleSubmit"
      @update:mentions="onUpdateMentions"
    />
    <!--      @update:hashtags="onUpdateHashtags"-->
  </section>
</template>
