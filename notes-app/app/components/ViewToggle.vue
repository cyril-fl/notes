<script setup lang="ts">
const view = useView();
</script>

<template>
  <UTooltip :text="view === 'grid' ? 'Vue liste' : 'Vue mosaïque'">
    <UButton
      :icon="view === 'grid' ? 'i-lucide-list' : 'i-lucide-layout-grid'"
      color="neutral"
      variant="ghost"
      :aria-label="view === 'grid' ? 'Vue liste' : 'Vue mosaïque'"
      @click="view = view === 'grid' ? 'list' : 'grid'"
    />
  </UTooltip>
</template>
