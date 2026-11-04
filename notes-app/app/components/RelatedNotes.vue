<script setup lang="ts">
import type { Neighbors } from '~/types/items';

const props = defineProps<{ noteId: string }>();

const api = useApi();
const { data } = useLazyAsyncData(`neighbors:${props.noteId}`, () =>
  api<Neighbors>(`/api/notes/${props.noteId}/neighbors`, { quiet: true })
);
</script>

<template>
  <section v-if="data?.neighbors.length" class="mt-10 border-t border-default pt-4">
    <div class="mb-2 flex items-center justify-between">
      <h2 class="text-xs font-medium text-muted">Notes liées</h2>
      <NuxtLink :to="{ path: '/graph', query: { note: noteId } }" class="text-xs text-primary">Voir le graphe</NuxtLink>
    </div>
    <NoteList :notes="data.neighbors" />
  </section>
</template>
