<script setup lang="ts">
const { icons } = useIcons();

const view = useView();
</script>

<template>
  <UTooltip :text="view === 'grid' ? $t('folder.listView') : $t('folder.gridView')">
    <UButton
      :icon="view === 'grid' ? icons.list : icons.grid"
      color="neutral"
      variant="ghost"
      :aria-label="view === 'grid' ? $t('folder.listView') : $t('folder.gridView')"
      @click="view = view === 'grid' ? 'list' : 'grid'"
    />
  </UTooltip>
</template>
