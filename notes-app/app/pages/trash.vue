<script setup lang="ts">
import type { Trashed } from '~/types/items';

const api = useApi();
const { restoreItem, emptyTrash } = useItems();
const { confirm } = useDialogs();

const { data: items, status } = useLazyAsyncData('trash', () => api<Trashed[]>('/api/trash'), {
  default: () => [] as Trashed[],
});

async function askEmpty() {
  const confirmed = await confirm({
    title: 'Vider la corbeille ?',
    description: 'Tout ce qu\'elle contient sera supprimé définitivement.',
    action: 'Vider',
  });
  if (confirmed) await emptyTrash();
}

function deletedOn(item: Trashed) {
  return new Date(item.deleted_at).toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' });
}
</script>

<template>
  <div class="mx-auto max-w-3xl p-8">
    <PageHeader title="Corbeille">
      <UButton v-if="items.length" label="Vider" color="error" variant="ghost" @click="askEmpty" />
    </PageHeader>
    <ul v-if="items.length">
      <li v-for="item in items" :key="item.id" class="row h-10 border-b border-muted last:border-0">
        <UIcon :name="item.kind === 'folder' ? 'i-lucide-folder' : 'i-lucide-file-text'" class="size-4 shrink-0 text-muted" />
        <span class="truncate">{{ displayTitle(item.title) }}</span>
        <span v-if="item.inside" class="shrink-0 text-xs text-muted">+{{ item.inside }}</span>
        <span class="ml-auto shrink-0 pl-4 text-xs text-muted">{{ deletedOn(item) }}</span>
        <UButton label="Restaurer" color="neutral" variant="ghost" size="xs" @click="restoreItem(item.id)" />
      </li>
    </ul>
    <p v-else-if="status !== 'pending'" class="py-16 text-center text-sm text-muted">La corbeille est vide.</p>
  </div>
</template>
