<script setup lang="ts">
// Same page for every note: a new note getting its address must keep its editor, and the cursor in it.
definePageMeta({ key: 'note' });

const route = useRoute();

function editorFor(key: number) {
  const id = route.params.id as string;
  return { key, noteId: id === 'new' ? undefined : id, parentId: route.query.parent as string | undefined };
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
  <NoteEditor :key="editor.key" :note-id="editor.noteId" :parent-id="editor.parentId" @created="onCreated" />
</template>
