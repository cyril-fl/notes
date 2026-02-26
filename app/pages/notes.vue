<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
const { t } = useI18n();

usePageSection({ title: t('pages.notes.title') });
</script>

<template>
  <ul class="space-y-4 flex flex-wrap gap-4">
    <li v-for="note in data.notes" :key="note.id">
      <UINotesCard :item="note" />
    </li>
  </ul>
</template>
