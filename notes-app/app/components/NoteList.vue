<script setup lang="ts">
defineProps<{ notes: { id: string; title: string; path: string; snippet?: string }[]; grid?: boolean }>();
</script>

<template>
  <ul v-if="grid" class="mosaic">
    <li v-for="note in notes" :key="note.id" class="min-w-0">
      <NuxtLink :to="`/notes/${note.id}`" class="group flex flex-col gap-2">
        <div class="tile items-start justify-start text-muted">
          <p class="line-clamp-6">{{ note.snippet ?? note.path }}</p>
        </div>
        <p class="truncate text-center text-xs">{{ displayTitle(note.title, $t('untitled')) }}</p>
      </NuxtLink>
    </li>
  </ul>
  <ul v-else>
    <li v-for="note in notes" :key="note.id">
      <NuxtLink
        :to="`/notes/${note.id}`"
        class="block rounded-md px-2 py-2 transition-colors duration-150 hover:bg-accented/60"
      >
        <p class="truncate text-sm">{{ displayTitle(note.title, $t('untitled')) }}</p>
        <p class="truncate text-xs text-muted">{{ note.snippet ?? note.path }}</p>
      </NuxtLink>
    </li>
  </ul>
</template>
