<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
const { t } = useI18n();

const sortedNotes = computed(() =>
  [...data.value.notes].sort((a, b) => {
    if (a.isPinned === b.isPinned) return 0;
    return a.isPinned ? -1 : 1;
  })
);

usePageSection({ title: t('pages.notes.title') });
</script>

<template>
  <ul class="space-y-4 flex flex-wrap gap-4">
    <li v-for="note in sortedNotes" :key="note.id">
      <UINotesCard :item="note" />
    </li>
  </ul>
</template>
