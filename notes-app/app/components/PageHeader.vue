<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>();
</script>

<template>
  <header class="mb-6 flex items-start gap-4">
    <div class="min-w-0 flex-1">
      <p v-if="subtitle" class="truncate text-xs text-muted">{{ subtitle }}</p>
      <h1 class="truncate text-2xl font-bold tracking-tight">{{ title }}</h1>
    </div>
    <div class="flex shrink-0 items-center gap-1"><slot /></div>
  </header>
</template>
