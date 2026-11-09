<script setup lang="ts">
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';
import type { ItemDetail } from '~/types/items';

const props = defineProps(nodeViewProps);
const { icons } = useIcons();
const api = useApi();
const baseURL = useRuntimeConfig().public.apiBase;

const id = computed<string | null>(() => props.node.attrs.id);
const target = computed<string>(() => props.node.attrs.target);
// Its embeds stay links: a note that includes itself, or one that includes it back, stops there.
const extensions = [WikiLink, noteEmbed(), noteImage(baseURL, async () => null)];

const content = ref<string>();
const missing = ref(false);
watch(
  id,
  async (noteId) => {
    content.value = undefined;
    missing.value = false;
    if (!noteId) return;
    const note = await api<ItemDetail>(`/api/notes/${noteId}`, { quiet: true }).catch(() => undefined);
    if (id.value !== noteId) return;
    content.value = note?.content ?? '';
    missing.value = !note;
  },
  { immediate: true },
);
</script>

<template>
  <NodeViewWrapper as="span" class="note-embed" :class="{ 'ring-2 ring-primary/60': selected }" contenteditable="false">
    <!-- The click handler of the editor opens it, like a [[link]]. -->
    <a
      :data-wikilink="target"
      :data-note="id ?? undefined"
      class="wikilink flex items-center gap-1.5 text-sm"
      :class="{ 'wikilink-waiting': !id || missing }"
    >
      <UIcon :name="missing ? icons.noteNotFound : icons.note" class="size-4 shrink-0" />
      <span class="truncate">{{ target }}</span>
    </a>
    <UEditor
      v-if="content"
      :model-value="content"
      content-type="markdown"
      :editable="false"
      :image="false"
      :mention="false"
      :extensions="extensions"
      :ui="{ base: 'px-0 sm:px-0' }"
    />
  </NodeViewWrapper>
</template>
