<script setup lang="ts">
defineProps<{ notes: { id: string; title: string; path: string; snippet?: string }[] }>();
</script>

<template>
  <ul>
    <li v-for="note in notes" :key="note.id" class="border-b border-muted last:border-0">
      <NuxtLink :to="`/notes/${note.id}`" class="block rounded-md px-2 py-2 transition-colors duration-150 hover:bg-accented/60">
        <p class="truncate text-sm">{{ displayTitle(note.title) }}</p>
        <p class="truncate text-xs text-muted">{{ note.snippet ?? note.path }}</p>
      </NuxtLink>
    </li>
  </ul>
</template>
