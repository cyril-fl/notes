<script setup lang="ts">
import type { User } from '~/types/items';

// Same page for every note: a new note getting its address must keep its editor, and the cursor in it.
definePageMeta({ key: 'note' });

const route = useRoute();
const api = useApi();
const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));

function editorFor(key: number) {
  const id = route.params.id as string;
  // shortcut: a reload of /notes/new?template= before me is loaded starts empty, wait for me if that bites.
  const template = id === 'new' ? me.value?.snippets[Number(route.query.template)]?.content : undefined;
  return { key, noteId: id === 'new' ? undefined : id, parentId: route.query.parent as string | undefined, template };
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
  <NoteEditor
    :key="editor.key"
    :note-id="editor.noteId"
    :parent-id="editor.parentId"
    :template="editor.template"
    @created="onCreated"
  />
</template>
