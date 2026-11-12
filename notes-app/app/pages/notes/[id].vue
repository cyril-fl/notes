<script setup lang="ts">
import type { User } from '~/types/items';

// Same page for every note: a new note getting its address must keep its editor, and the cursor in it.
definePageMeta({ key: 'note' });

const route = useRoute();
const api = useApi();
const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));

function editorFor(key: number) {
  const id = route.params.id as string;
  // TODO: a reload of /notes/new?snippet= before me is loaded starts empty, wait for me if that bites.
  const start = id === 'new' ? me.value?.snippets[Number(route.query.snippet)]?.content : undefined;
  return { key, noteId: id === 'new' ? undefined : id, parentId: route.query.parent as string | undefined, start };
}

const editor = ref(editorFor(0));
let createdId: string | undefined;

watch(
  () => route.fullPath,
  () => {
    if (route.params.id !== createdId) editor.value = editorFor(editor.value.key + 1);
    createdId = undefined;
  },
);

function onCreated(id: string) {
  createdId = id;
  navigateTo(`/notes/${id}`, { replace: true });
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <NoteTabs />
    <NoteEditor
      :key="editor.key"
      :note-id="editor.noteId"
      :parent-id="editor.parentId"
      :start="editor.start"
      @created="onCreated"
    />
  </div>
</template>
