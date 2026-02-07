<script setup lang="ts">
import twitter from 'twitter-text';
import striptags from 'striptags';

/* Define */
export interface EditorProps {
  readonly: boolean;
  showHashtags: boolean;
  showMentions: boolean;
  showPreview: boolean;
}

interface Emits {
  (e: 'submit'): void;
  (e: 'update:hashtags' | 'update:mentions', value: string[]): void;
}

const props = withDefaults(defineProps<Partial<EditorProps>>(), {
  readonly: false,
  showHashtags: true,
  showMentions: true,
  showPreview: true,
});

const emit = defineEmits<Emits>();

const content = defineModel<string | null>('content');

/* Data */
const hashtags = computed(() => twitter.extractHashtags(content.value ?? ''));

const mentions = computed(() => twitter.extractMentions(content.value ?? ''));

const preview = computed(() => [
  {
    label: 'Preview',
    condition: props.showPreview,
    content: content.value ?? '  ',
  },
  {
    label: 'Extracted hashtags',
    condition: props.showHashtags,
    content: hashtags.value,
  },
  {
    label: 'Extracted mentions',
    condition: props.showMentions,
    content: mentions.value,
  },
]);

/* Methods */
const handleSubmit = () => {
  emit('submit');
};

const handleUpdateContent = (newContent: string | null) => {
  if (newContent === null || newContent === undefined) {
    content.value = null;
    return;
  }
  const normalized = striptags(newContent)
    .replace(/&nbsp;/gi, ' ')
    .replace(/\u00A0/g, ' ')
    .trim();
  content.value = normalized.length !== 0 ? normalized : null;
};

/* Lifecycle */
watch(hashtags, (newHashtags, oldHashtags) => {
  if (!oldHashtags || oldHashtags.every((tag) => newHashtags.includes(tag)))
    return;
  emit('update:hashtags', newHashtags);
});

watch(mentions, (newMentions, oldMentions) => {
  if (
    !oldMentions ||
    oldMentions.every((mention) => newMentions.includes(mention))
  )
    return;
  emit('update:mentions', newMentions);
});

onBeforeRouteLeave(handleSubmit);

onMounted(() => {
  window.addEventListener('beforeunload', handleSubmit);
});

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleSubmit);
});
</script>

<template>
  <ClientOnly>
    <EditorCore
      :content="content"
      :placeholder="$t('placeholder.editor')"
      :readonly="readonly"
      @update:content="handleUpdateContent"
    />
    <template #fallback>
      <div
        class="bg-muted min-h-21 w-full rounded-md border border-default px-3 py-2"
      >
        <span class="text-muted">{{ $t('placeholder.editor') }}</span>
      </div>
    </template>
  </ClientOnly>

  <ul class="mt-4">
    <template v-for="item in preview" :key="item.label">
      <li v-if="item.condition">
        <p>{{ item.label }} :</p>
        <pre>{{ item.content }}</pre>
      </li>
    </template>
  </ul>
</template>
