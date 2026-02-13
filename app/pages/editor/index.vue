<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
</script>

<template>
  <section class="space-y-8">
    <header>
      <p>Liste des notes</p>
    </header>

    <ul class="space-y-4 flex flex-wrap gap-4">
      <li v-for="note in data.notes" :key="note.id">
        <UINotesCard :note="note" />
      </li>
    </ul>
  </section>
</template>
