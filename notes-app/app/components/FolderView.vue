<script setup lang="ts">
import type { Item } from '~/types/items';

const { icons } = useIcons();

const props = defineProps<{ folderId?: string; title: string; subtitle?: string; items: Item[] }>();

const { createFolder, moveItem, removeItems } = useItems();
const view = useView();

// ⌘-click (Ctrl elsewhere) adds or takes out an item, ⇧-click selects a range from the last one; a plain click opens it.
const selection = useSelection();
const selected = computed(() => props.items.filter((item) => selection.value.includes(item.id)));
let anchor = -1;

function select(event: MouseEvent, index: number) {
  if (!event.metaKey && !event.ctrlKey && !event.shiftKey) return;
  event.preventDefault();
  event.stopPropagation(); // The link under it must not open.
  const id = props.items[index]!.id;
  if (event.shiftKey && anchor >= 0) {
    const range = props.items.slice(Math.min(anchor, index), Math.max(anchor, index) + 1).map((item) => item.id);
    selection.value = [...new Set([...selection.value, ...range])];
    return;
  }
  selection.value = selection.value.includes(id) ? selection.value.filter((other) => other !== id) : [...selection.value, id];
  anchor = index;
}

onKeyStroke('Escape', () => (selection.value = []));

// A selected folder cannot go into itself: it is not offered.
const destinations = useDestinations(() => selection.value);

async function moveSelection(destination: string) {
  await moveItem(selection.value, destination === ROOT ? null : destination);
  selection.value = [];
}

async function removeSelection() {
  await removeItems(selected.value);
  selection.value = [];
}

const panel = useTemplateRef('panel');
const isOver = useDropFolder(panel, () => props.folderId ?? null);
</script>

<template>
  <div ref="panel" class="flex min-h-full flex-col p-8 transition-shadow duration-150" :class="{ 'ring-2 ring-inset ring-primary/40': isOver }">
    <PageHeader :title="title">
      <ViewToggle />
      <UTooltip text="Nouveau dossier">
        <UButton :icon="icons.folderAdd" color="neutral" variant="ghost" aria-label="Nouveau dossier" @click="createFolder(folderId)" />
      </UTooltip>
      <UTooltip text="Nouvelle note">
        <UButton :icon="icons.noteAdd" color="neutral" variant="ghost" aria-label="Nouvelle note" :to="newNotePath(folderId)" />
      </UTooltip>
    </PageHeader>
    <div v-if="selected.length" class="sticky top-0 z-10 mb-3 flex items-center gap-2 rounded-md bg-elevated py-1 pr-1 pl-3 text-sm">
      <span class="mr-auto">{{ selected.length }} sélectionné{{ selected.length > 1 ? 's' : '' }}</span>
      <USelectMenu
        :items="destinations"
        value-key="value"
        placeholder="Déplacer vers…"
        size="sm"
        class="w-56"
        aria-label="Déplacer vers"
        @update:model-value="moveSelection"
      />
      <UButton label="Supprimer" :icon="icons.delete" color="error" variant="ghost" size="sm" @click="removeSelection" />
      <UTooltip text="Annuler la sélection" :kbds="['esc']">
        <UButton :icon="icons.close" color="neutral" variant="ghost" size="sm" aria-label="Annuler la sélection" @click="selection = []" />
      </UTooltip>
    </div>
    <ul v-if="items.length" :class="{ mosaic: view === 'grid' }">
      <li v-for="(item, index) in items" :key="item.id" @click.capture="select($event, index)">
        <ItemRow :item="item" :grid="view === 'grid'" :selected="selection.includes(item.id)" />
      </li>
    </ul>
    <UEmpty
      v-else
      :icon="icons.folderEmpty"
      title="Ce dossier est vide"
      description="Créez une note ou glissez-y des éléments depuis la barre latérale."
      :actions="[
        { label: 'Nouvelle note', icon: icons.noteAdd, to: newNotePath(folderId) },
        { label: 'Nouveau dossier', icon: icons.folderAdd, color: 'neutral', variant: 'subtle', onClick: () => createFolder(folderId) },
      ]"
    />
    <!-- Like a note's footer. The subtitle is the folder's own path: the breadcrumb leads to its parents. -->
    <footer v-if="subtitle" class="sticky bottom-0 z-10 -mx-8 -mb-8 mt-auto border-t border-default bg-default px-8">
      <div class="flex h-9 items-center text-xs text-muted">
        <PathNav :path="subtitle.slice(0, subtitle.lastIndexOf('/'))" />
      </div>
    </footer>
  </div>
</template>
