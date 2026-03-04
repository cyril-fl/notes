<script setup lang="ts">
const store = useDataStore();
const { data } = storeToRefs(store);
const { t } = useI18n();
const { clearSelection } = useSelection();

// Main view as drop zone for root
const contentRef = useTemplateRef('contentRef');
const rootFolderId = ref<string | null>(null);
const { isOver: isDropOver } = useDropZone({ elementRef: contentRef, folderId: rootFolderId });

const sortedNotes = computed(() =>
  [...data.value.notes].sort((a, b) => {
    if (a.isPinned === b.isPinned) return 0;
    return a.isPinned ? -1 : 1;
  })
);

usePageSection({ title: t('pages.notes.title') });
</script>

<template>
  <div
    ref="contentRef"
    class="grow transition-all"
    :class="{ 'ring-2 ring-primary/30 rounded-lg': isDropOver }"
  >
    <ul class="space-y-4 flex flex-wrap gap-4" @click.self="clearSelection()">
      <li v-for="note in sortedNotes" :key="note.id">
        <UINotesCard :item="note" />
      </li>
    </ul>
  </div>
</template>
