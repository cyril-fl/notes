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
    <h2 class="mb-2 text-xs font-medium text-muted">Notes liées</h2>
    <NoteList :notes="data.neighbors" />
  </section>
</template>
