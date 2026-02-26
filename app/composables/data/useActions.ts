import type { ContextMenuItem } from '@nuxt/ui';
import type { Ref } from 'vue';

export const useActions = (
  folderId: Ref<string | null>,
  { requestEdit }: { requestEdit: (id: string) => void }
) => {
  const { getById } = useDataUtils();
  const { createFolderInFolder, deleteById } = useDataActions();
  const { emitFolderCardUpdate } = useUIEvents();
  const { t } = useI18n();
  const { icons } = useIcons();

  const isRoot = computed(() => folderId.value === 'root');

  const addFolder = (
    callback?: (newId: string, parentId: string) => void | Promise<void>
  ): ContextMenuItem => ({
    label: t('menu.context.add_folder'),
    icon: icons.folderadd,
    onSelect: async () => {
      const parentId = folderId.value;
      if (!parentId) return;
      const folder = getById(parentId, { types: ItemType.FOLDER });
      if (!folder) return;
      const newFolder = await createFolderInFolder({ folder });
      if (!newFolder) return;
      if (callback) await callback(newFolder.id, parentId);
    },
  });

  const addNote = (): ContextMenuItem => ({
    label: t('menu.context.add_note'),
    icon: icons.noteadd,
    onSelect: () => {
      const id = folderId.value;
      if (!id) return;
      navigateTo(`${NAVIGATION.newNote}${id}/`);
    },
  });

  const deleteItem = (): ContextMenuItem => ({
    label: t('menu.context.delete'),
    icon: icons.delete,
    color: 'error',
    disabled: isRoot.value,
    onSelect: async () => {
      const id = folderId.value;
      if (!id) return;

      const item = getById(id);
      if (!item) return;

      const itemType = item.type === ItemType.FOLDER ? 'folder' : 'note';
      const { requestConfirm } = useConfirmDelete();
      const confirmed = await requestConfirm(id, itemType);

      if (confirmed) {
        deleteById(id);
      }
    },
  });

  const updateItem = (): ContextMenuItem => ({
    label: t('menu.context.update'),
    icon: icons.edit,
    onSelect: () => {
      const id = folderId.value;
      if (!id) return;
      requestEdit(id);
    },
  });

  const updateFolderCard = (): ContextMenuItem => ({
    label: t('menu.context.rename'),
    icon: icons.edit,
    onSelect: () => {
      const id = folderId.value;
      if (!id) return;
      emitFolderCardUpdate(id);
    },
    disabled: isRoot.value,
  });

  const updateFolderLink = (): ContextMenuItem => ({
    label: t('menu.context.rename'),
    icon: icons.edit,
    onSelect: () => {
      const id = folderId.value;
      if (!id) return;
      requestEdit(id);
    },
    disabled: isRoot.value,
  });

  return {
    addFolder,
    addNote,
    deleteItem,
    updateItem,
    updateFolderCard,
    updateFolderLink,
  };
};
