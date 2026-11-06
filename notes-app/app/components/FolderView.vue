<script setup lang="ts">
import type { Item } from '~/types/items';

const props = defineProps<{ folderId?: string; title: string; subtitle?: string; items: Item[] }>();

const { createFolder } = useItems();
const view = useView();

const panel = useTemplateRef('panel');
const isOver = useDropFolder(panel, () => props.folderId ?? null);
</script>

<template>
  <div ref="panel" class="min-h-full p-8 transition-shadow duration-150" :class="{ 'ring-2 ring-inset ring-primary/40': isOver }">
    <PageHeader :title="title" :subtitle="subtitle">
      <ViewToggle />
      <UTooltip text="Nouveau dossier">
        <UButton icon="i-lucide-folder-plus" color="neutral" variant="ghost" aria-label="Nouveau dossier" @click="createFolder(folderId)" />
      </UTooltip>
      <UTooltip text="Nouvelle note">
        <UButton icon="i-lucide-square-pen" color="neutral" variant="ghost" aria-label="Nouvelle note" :to="newNotePath(folderId)" />
      </UTooltip>
    </PageHeader>
    <ul v-if="items.length" :class="{ mosaic: view === 'grid' }">
      <li v-for="item in items" :key="item.id" :class="{ 'border-b border-muted last:border-0': view === 'list' }">
        <ItemRow :item="item" :grid="view === 'grid'" />
      </li>
    </ul>
    <UEmpty
      v-else
      icon="i-lucide-folder-open"
      title="Ce dossier est vide"
      description="Créez une note ou glissez-y des éléments depuis la barre latérale."
      :actions="[
        { label: 'Nouvelle note', icon: 'i-lucide-square-pen', to: newNotePath(folderId) },
        { label: 'Nouveau dossier', icon: 'i-lucide-folder-plus', color: 'neutral', variant: 'subtle', onClick: () => createFolder(folderId) },
      ]"
    />
  </div>
</template>
