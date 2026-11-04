<script setup lang="ts">
import type { Item } from '~/types/items';

const props = defineProps<{ folderId?: string; title: string; subtitle?: string; items: Item[] }>();

const { createFolder } = useItems();

const panel = useTemplateRef('panel');
const isOver = useDropFolder(panel, () => props.folderId ?? null);
</script>

<template>
  <div ref="panel" class="min-h-full p-8 transition-shadow duration-150" :class="{ 'ring-2 ring-inset ring-primary/40': isOver }">
    <div class="mx-auto max-w-3xl">
      <PageHeader :title="title" :subtitle="subtitle">
        <slot name="actions" />
        <UTooltip text="Nouveau dossier">
          <UButton icon="i-lucide-folder-plus" color="neutral" variant="ghost" aria-label="Nouveau dossier" @click="createFolder(folderId)" />
        </UTooltip>
        <UTooltip text="Nouvelle note">
          <UButton icon="i-lucide-square-pen" color="neutral" variant="ghost" aria-label="Nouvelle note" :to="newNotePath(folderId)" />
        </UTooltip>
      </PageHeader>
      <ul v-if="items.length">
        <li v-for="item in items" :key="item.id" class="border-b border-muted last:border-0">
          <ItemRow :item="item" />
        </li>
      </ul>
      <p v-else class="py-16 text-center text-sm text-muted">Ce dossier est vide.</p>
    </div>
  </div>
</template>
