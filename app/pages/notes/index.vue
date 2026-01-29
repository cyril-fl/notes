<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
const { handleDeleteId } = useDataApi();

const handleDelete = (note: Note) => {
  handleDeleteId(note.id);
};
</script>

<template>
  <section class="space-y-8">
    <header>
      <p>Liste des notes</p>
    </header>

    <ul class="space-y-4">
      <li v-for="note in data.notes" :key="note.id">
        <NuxtLink :to="`/notes/${note.id}`" as="div">
          <p>id: {{ note.id }}</p>
          <p>title: {{ note.title }}</p>
          <p>content: {{ note.content }}</p>
          <p>hashtags: {{ note.hashtags }}</p>
          <p>mentions: {{ note.mentions }}</p>
          <p>createdAt: {{ note.createdAt }}</p>
          <p>updatedAt: {{ note.updatedAt }}</p>
        </NuxtLink>
        <button
          class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
          @click="() => handleDelete(note)"
        >
          DELETE
        </button>
      </li>
    </ul>
  </section>
</template>
