<script setup lang="ts">
import type { ItemDetail } from '~/types/items';

const route = useRoute();
const api = useApi();
const { renameFolder, removeItem } = useItems();

const id = route.params.id as string;
const { data: folder, error } = useLazyAsyncData(`folder:${id}`, () => api<ItemDetail>(`/api/folders/${id}`, { quiet: true }));

async function remove() {
  if (!folder.value) return;
  const { item } = folder.value;
  // Leave first: the views read again after a removal, and this one would ask for a folder that is gone.
  await navigateTo(item.parent_id ? `/folders/${item.parent_id}` : '/');
  await removeItem(item);
}
</script>

<template>
  <FolderView v-if="folder" :folder-id="id" :title="folder.item.title" :subtitle="folder.path" :items="folder.children ?? []">
    <template #actions>
      <UTooltip text="Renommer">
        <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" aria-label="Renommer le dossier" @click="renameFolder(folder!.item)" />
      </UTooltip>
      <UTooltip text="Supprimer">
        <UButton icon="i-lucide-trash-2" color="neutral" variant="ghost" aria-label="Supprimer le dossier" @click="remove" />
      </UTooltip>
    </template>
  </FolderView>
  <p v-else-if="error" class="py-24 text-center text-sm text-muted">Ce dossier n'existe pas ou est dans la corbeille.</p>
</template>
