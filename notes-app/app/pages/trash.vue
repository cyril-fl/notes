<script setup lang="ts">
import type { Trashed } from '~/types/items';

const { icons } = useIcons();

const api = useApi();
const { restoreItem, emptyTrash, purgeItem } = useItems();
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

async function askPurge(item: Trashed) {
  const confirmed = await confirm({
    title: `Supprimer « ${displayTitle(item.title)} » définitivement ?`,
    description: item.inside ? `Les ${item.inside} éléments supprimés avec lui le seront aussi.` : 'Il ne pourra plus être restauré.',
    action: 'Supprimer',
  });
  if (confirmed) await purgeItem(item.id);
}

function deletedOn(item: Trashed) {
  return new Date(item.deleted_at).toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' });
}
</script>

<template>
  <div class="p-8">
    <PageHeader title="Corbeille">
      <UButton v-if="items.length" label="Vider" color="error" variant="ghost" @click="askEmpty" />
    </PageHeader>
    <ul v-if="items.length">
      <li v-for="item in items" :key="item.id" class="row h-10 border-b border-muted last:border-0">
        <UIcon :name="item.kind === 'folder' ? icons.folder : icons.note" class="size-4 shrink-0 text-muted" />
        <span class="truncate">{{ displayTitle(item.title) }}</span>
        <span v-if="item.inside" class="shrink-0 text-xs text-muted">+{{ item.inside }}</span>
        <span class="ml-auto shrink-0 pl-4 text-xs text-muted">{{ deletedOn(item) }}</span>
        <UButton label="Restaurer" color="neutral" variant="ghost" size="xs" @click="restoreItem(item.id)" />
        <UTooltip text="Supprimer définitivement">
          <UButton :icon="icons.close" color="neutral" variant="ghost" size="xs" aria-label="Supprimer définitivement" @click="askPurge(item)" />
        </UTooltip>
      </li>
    </ul>
    <UEmpty
      v-else-if="status !== 'pending'"
      :icon="icons.trash"
      title="La corbeille est vide"
      description="Les notes et les dossiers supprimés arrivent ici, et peuvent en être restaurés."
    />
  </div>
</template>
