<script setup lang="ts">
defineProps<{
  notes: { id: string; title: string; path: string; snippet?: string; readonly?: boolean }[];
  grid?: boolean;
}>();

const { icons } = useIcons();
</script>

<template>
  <ul v-if="grid" class="mosaic">
    <li v-for="note in notes" :key="note.id" class="min-w-0">
      <NuxtLink :to="`/notes/${note.id}`" class="group flex flex-col gap-2">
        <div class="tile relative items-start justify-start text-muted">
          <p class="line-clamp-6">{{ note.snippet ?? note.path }}</p>
          <UIcon
            v-if="note.readonly"
            :name="icons.lock"
            :aria-label="$t('note.readonly')"
            class="absolute top-2 right-2 size-4 text-muted"
          />
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
        <p class="flex items-center gap-2 text-sm">
          <span class="truncate">{{ displayTitle(note.title, $t('untitled')) }}</span>
          <UIcon
            v-if="note.readonly"
            :name="icons.lock"
            :aria-label="$t('note.readonly')"
            class="size-3.5 shrink-0 text-muted"
          />
        </p>
        <p class="truncate text-xs text-muted">{{ note.snippet ?? note.path }}</p>
      </NuxtLink>
    </li>
  </ul>
</template>
