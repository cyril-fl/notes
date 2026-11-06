<script setup lang="ts">
import type { ItemDetail } from '~/types/items';

const route = useRoute();
const api = useApi();

const id = route.params.id as string;
const { data: folder, error } = useLazyAsyncData(`folder:${id}`, () => api<ItemDetail>(`/api/folders/${id}`, { quiet: true }));
</script>

<template>
  <FolderView v-if="folder" :folder-id="id" :title="folder.item.title" :subtitle="folder.path" :items="folder.children ?? []" />
  <UEmpty
    v-else-if="error"
    icon="i-lucide-folder-x"
    title="Ce dossier est introuvable"
    description="Il n'existe pas ou il est dans la corbeille."
    :actions="[{ label: 'Retour aux notes', icon: 'i-lucide-arrow-left', color: 'neutral', variant: 'subtle', to: '/' }]"
  />
</template>
