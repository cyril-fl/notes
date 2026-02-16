<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
</script>

<template>
  <!-- TODO add context menue -->
  <UIPageSection :title="$t('pages.editor.title')">
    <ul class="space-y-4 flex flex-wrap gap-4">
      <li v-for="note in data.notes" :key="note.id">
        <UINotesCard :item="note" />
      </li>
    </ul>
  </UIPageSection>
</template>
