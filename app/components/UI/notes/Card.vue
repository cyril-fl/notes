<script setup lang="ts">
// Define
interface NotesCardProps {
  note: Note;
}

defineProps<NotesCardProps>();
// Data

// Methods

// Lifecycle

// SEO
</script>

<template>
  <div class="flex flex-col gap-2">
    <NuxtLink
      :to="`${NAVIGATION.notes}${note.id}`"
      as="p"
      class="bg-muted text-xs text-default p-4 rounded-md aspect-square size-30 overflow-hidden"
    >
    <span class="line-clamp-5">
      {{ note.content }}
    </span>
    </NuxtLink>
    <ul class="text-xs text-muted text-center">
      <li>{{ note.updatedAt }}</li>
    </ul>
    <button @click="() => $hooks.callHook('data:delete:id', note.id)">DELETE</button>
  </div>
</template>
