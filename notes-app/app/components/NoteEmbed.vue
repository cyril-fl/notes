<script setup lang="ts">
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';
import type { ItemDetail } from '~/types/items';

const props = defineProps(nodeViewProps);
const { icons } = useIcons();
const api = useApi();
const baseURL = useRuntimeConfig().public.apiBase;

const id = computed<string | null>(() => props.node.attrs.id);
const target = computed<string>(() => props.node.attrs.target);
const heading = computed<string | null>(() => props.node.attrs.heading);
// The tables and callouts every editor reads, without their menus: this one is read only.
const [tables] = useNoteTables();
const [callouts] = useNoteCallouts();
// Its embeds stay links: a note that includes itself, or one that includes it back, stops there.
const extensions = [
  WikiLink,
  noteEmbed(),
  noteImage(baseURL, async () => null),
  ...tables!.extensions!,
  ...callouts!.extensions!,
];

const text = ref<string>();
const missing = ref(false);
const content = computed(
  () => text.value && (heading.value ? sectionOf(text.value, heading.value) : withoutTitle(text.value)),
);
watch(
  id,
  async (noteId) => {
    text.value = undefined;
    missing.value = false;
    if (!noteId) return;
    const note = await api<ItemDetail>(`/api/notes/${noteId}`, { quiet: true }).catch(() => undefined);
    if (id.value !== noteId) return;
    text.value = note?.content ?? '';
    missing.value = !note;
  },
  { immediate: true },
);

// Alone in its paragraph it shows the note; in a sentence, a link like a [[link]].
const alone = ref(true);
function place() {
  const pos = props.getPos();
  if (pos === undefined) return;
  const parent = props.editor.state.doc.resolve(pos).parent;
  alone.value = parent.content.content.filter((child) => !child.isText || child.text?.trim()).length === 1;
}
onMounted(() => {
  place();
  props.editor.on('update', place);
});
onBeforeUnmount(() => props.editor.off('update', place));
</script>

<template>
  <NodeViewWrapper
    as="span"
    :class="{ 'note-embed': alone, 'ring-2 ring-primary/60': selected }"
    contenteditable="false"
  >
    <!-- The click handler of the editor opens it, like a [[link]]. -->
    <a
      :data-wikilink="target"
      :data-note="id ?? undefined"
      class="wikilink"
      :class="[alone ? 'flex items-center gap-1.5 text-sm' : 'inline', { 'wikilink-waiting': !id || missing }]"
    >
      <UIcon v-if="alone" :name="missing ? icons.noteNotFound : icons.note" class="size-4 shrink-0" />
      <span class="truncate">{{ target }}</span>
    </a>
    <UEditor
      v-if="alone && content"
      :model-value="content"
      content-type="markdown"
      :editable="false"
      :image="false"
      :mention="false"
      :starter-kit="callouts!.starterKit"
      :extensions="extensions"
      :ui="{ base: 'px-0 sm:px-0' }"
    />
  </NodeViewWrapper>
</template>
