<script setup lang="ts">
import type { ItemDetail } from '~/types/items';

// Same page for every note: a new note getting its address must keep its editor, and the cursor in it.
definePageMeta({ key: 'note' });

const route = useRoute();
const api = useApi();
const { data: me } = useMe();

function editorFor() {
  const id = route.params.id as string;
  // TODO: a reload of /notes/new?snippet= before me is loaded starts empty, wait for me if that bites.
  const start = id === 'new' ? me.value?.snippets[Number(route.query.snippet)]?.content : undefined;
  return { noteId: id === 'new' ? undefined : id, parentId: route.query.parent as string | undefined, start };
}

const editor = ref<ReturnType<typeof editorFor> & { key: number; note?: ItemDetail }>({ key: 0, ...editorFor() });
let createdId: string | undefined;
let opening = 0;

// The next note is read before its editor replaces the one shown, or the page blinks empty while it loads (#154).
// Not read, it loads itself and says why. Only the last note asked for opens.
watch(
  () => route.fullPath,
  async () => {
    const created = route.params.id === createdId;
    createdId = undefined;
    if (created) return;
    const turn = ++opening;
    const next = editorFor();
    const note = next.noteId
      ? await api<ItemDetail>(`/api/notes/${next.noteId}`, { quiet: true }).catch(() => undefined)
      : undefined;
    if (turn === opening) editor.value = { ...next, key: editor.value.key + 1, note };
  },
);
// The panel keeps the related notes and the graph from one note to the next, not for a note opened long after.
onUnmounted(() => clearNuxtState(['related-shown', 'graph-shown']));

function onCreated(id: string) {
  createdId = id;
  navigateTo(`/notes/${id}`, { replace: true });
}
</script>

<template>
  <NoteEditor
    :key="editor.key"
    :note-id="editor.noteId"
    :parent-id="editor.parentId"
    :start="editor.start"
    :note="editor.note"
    @created="onCreated"
  />
</template>
