<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';

const { folders } = useNavigationTree();

const { emitFolderNavOpen } = useUIEvents();
const { actions, editingId, clearEdit, reset } = useCurrentFolder();

const isMenuOpen = ref(false);

function whenMenuClosed(): Promise<void> {
  if (!isMenuOpen.value) return Promise.resolve();
  return new Promise((resolve) => {
    const stop = watch(isMenuOpen, (open) => {
      if (!open) {
        stop();
        resolve();
      }
    });
  });
}

const items = computed<ContextMenuItem[][]>(() => [
  [
    actions.addFolder(async (newId, parentId) => {
      await whenMenuClosed();
      emitFolderNavOpen(parentId);
      navigateTo(`/${newId}/?rename=1`);
    }),
    actions.addNote(),
  ],
  [actions.updateFolderLink()],
  [actions.deleteItem()],
]);

function handleCloseAutoFocus(event: Event) {
  if (editingId.value) event.preventDefault();
}

function handleMouseLeave() {
  if (!isMenuOpen.value) reset();
}

function handleOpenChange(open: boolean) {
  isMenuOpen.value = open;
  if (open) clearEdit();
  if (!open) reset();
}
</script>

<template>
  <div @mouseleave="handleMouseLeave">
    <UContextMenu
      :items="items"
      :content="{ onCloseAutoFocus: handleCloseAutoFocus }"
      @update:open="handleOpenChange"
    >
      <UINavigationTreeMenuNode
        v-for="folder in folders"
        :key="`folder-${folder.label}`"
        :item="folder"
      />
    </UContextMenu>
  </div>
</template>

<style scoped></style>
