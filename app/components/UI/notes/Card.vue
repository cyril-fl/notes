<script setup lang="ts">
// Define
const { handleDeleteId } = useDataApi();

const handleDelete = (note: Note) => {
  handleDeleteId(note.id);
};

interface NotesCardProps {
  note: Note;
}

defineProps<NotesCardProps>();
// Data

// Methods

// Lifecycle

// SEO
</script>

<template>
  <NuxtLink :to="`/notes/${note.id}`" as="div">
    <p>id: {{ note.id }}</p>
    <p>id: {{ note.id }}</p>
    <p>title: {{ note.title }}</p>
    <p>content: {{ note.content }}</p>
    <p>hashtags: {{ note.hashtags }}</p>
    <p>mentions: {{ note.mentions }}</p>
    <p>createdAt: {{ note.createdAt }}</p>
    <p>updatedAt: {{ note.updatedAt }}</p>
    <div class="bg-muted p-4 rounded-md w-full text-xs text-muted">
      <pre>{{ note }}</pre>
    </div>
    <button @click="() => handleDelete(note)">DELETE</button>
  </NuxtLink>
</template>

<style scoped></style>
