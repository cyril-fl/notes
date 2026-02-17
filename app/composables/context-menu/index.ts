import type { ContextMenuItem } from '@nuxt/ui';

export const useActions = () => {
  const { $hooks } = useNuxtApp();
  const { getById } = useDataUtils();
  const { t } = useI18n();
  const { icons } = useIcons();

  const addFolder = (
    folderId: string,
    callback?: (newId: string) => void | Promise<void>
  ): ContextMenuItem => ({
    label: t('menu.context.add_folder'),
    icon: icons.folderadd,
    onSelect: async () => {
      if (callback) {
        // Création directe + callback (ex: Card / page dossier)
        const folder = getById(folderId, { types: ItemType.FOLDER });
        if (!folder) return;

        const results = await $hooks.callHookParallel(
          'data:create:folder:in-folder',
          { folder }
        );

        const newFolder = Array.isArray(results)
          ? results.find((r) => r !== undefined)
          : undefined;
        if (!newFolder) return;

        await callback(newFolder.id);
      } else {
        // Pas de callback → ouvre la modale (ex: arbre de navigation)
        const { openCreateModal } = useFolder();
        openCreateModal(folderId);
      }
    },
  });

  const addNote = (folderId: string): ContextMenuItem => ({
    label: t('menu.context.add_note'),
    icon: icons.noteadd,
    to: `${NAVIGATION.newNote}${folderId}/`,
  });

  const deleteItem = (id: string): ContextMenuItem => ({
    label: t('menu.context.delete'),
    icon: icons.delete,
    color: 'error',
    onSelect: () => $hooks.callHook('data:delete:id', id),
  });

  const updateItem = (id: string): ContextMenuItem => ({
    label: t('menu.context.update'),
    icon: icons.edit,
    onSelect: () => $hooks.callHook('data:update:id', id),
  });

  const updateFolderCard = (id: string): ContextMenuItem => ({
    label: t('menu.context.rename'),
    icon: icons.edit,
    onSelect: () => $hooks.callHook('folder-card:update', id),
    disabled: id === 'root',
  });

  const updateFolderLink = (id: string): ContextMenuItem => {
    const { requestEdit: _requestEdit } = useFolder();
    return {
      label: t('menu.context.rename'),
      icon: icons.edit,
      onSelect: () => _requestEdit(id),
      disabled: id === 'root',
    };
  };

  return {
    addFolder,
    addNote,
    deleteItem,
    updateItem,
    updateFolderCard,
    updateFolderLink,
  };
};
