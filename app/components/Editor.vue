<script setup lang="ts">
import Emoji from '@tiptap/extension-emoji';
import TextAlign from '@tiptap/extension-text-align';
import Highlight from '@tiptap/extension-highlight';
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

defineProps<Partial<EditorProps>>();

const emit = defineEmits<Emits>();

const content = defineModel<string | null>('content');
const emojiItems = useEditorEmoji();
const mentionsItems = useEditorMention();
const toolBarItems = useEditorToolBar();

/* Data */
const extensions = [Emoji, TextAlign, Highlight];

const hashtags = computed(() =>
  content.value ? twitter.extractHashtags(content.value) : []
);
const mentions = computed(() =>
  content.value ? twitter.extractMentions(content.value) : []
);

// SSR-safe function to append menus to body (avoids z-index issues in docs)
// eslint-disable-next-line no-constant-condition
const appendToBody = false ? () => document.body : undefined;

/* Methods */
const handleSubmit = () => {
  emit('submit');
};

const handleUpdateContent = (newContent: string | null) => {
  if (!newContent) return;

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
// Save on navigation
onBeforeRouteLeave(handleSubmit);

// Save on close/refresh
onMounted(() => {
  window.addEventListener('beforeunload', handleSubmit);
});

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleSubmit);
});

/* SEO */
</script>

<template>
  <UEditor
    v-slot="{ editor }"
    v-model="content"
    :extensions="extensions"
    content-type="markdown"
    :placeholder="$t('components.editor.placeholder')"
    :editable="!readonly"
    class="bg-muted w-full min-h-21"
    @update:model-value="handleUpdateContent"
  >
    <UEditorToolbar v-if="!readonly" :editor="editor" :items="toolBarItems" />
    <UEditorEmojiMenu
      :editor="editor"
      :items="emojiItems"
      :append-to="appendToBody"
    />
    <UEditorMentionMenu
      :editor="editor"
      :items="mentionsItems"
      :append-to="appendToBody"
    />
  </UEditor>

  <p v-if="showPreview">
    Preview:
    {{ content }}
  </p>
  <p v-if="showHashtags">
    Hashtags:
    {{ hashtags }}
  </p>
  <p v-if="showMentions">
    Mentions:
    {{ mentions }}
  </p>
</template>
