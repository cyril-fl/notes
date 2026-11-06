<script setup lang="ts">
import type { Neighbors } from '~/types/items';

const props = defineProps<{ noteId: string }>();

const api = useApi();
const { data } = useLazyAsyncData(`neighbors:${props.noteId}`, () =>
  api<Neighbors>(`/api/notes/${props.noteId}/neighbors`, { quiet: true })
);

const sections = computed(() =>
  [
    { title: 'Liens', notes: data.value?.links ?? [] },
    { title: 'Rétroliens', notes: data.value?.backlinks ?? [] },
    { title: 'Mêmes tags', notes: data.value?.neighbors ?? [] },
  ].filter((section) => section.notes.length)
);
</script>

<template>
  <section v-if="sections.length" class="mt-10 space-y-6 border-t border-default pt-4">
    <div v-for="(section, index) in sections" :key="section.title">
      <div class="mb-2 flex items-center justify-between">
        <h2 class="text-xs font-medium text-muted">{{ section.title }}</h2>
        <NuxtLink v-if="index === 0" :to="{ path: '/graph', query: { note: noteId } }" class="text-xs text-primary">Voir le graphe</NuxtLink>
      </div>
      <NoteList :notes="section.notes" />
    </div>
  </section>
</template>
