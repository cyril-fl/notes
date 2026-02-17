interface FolderState {
  /** ID du dossier en cours d'édition inline (rename), ou undefined */
  editingId?: string;
  /** État de la modale de création de dossier */
  createModal: {
    isOpen: boolean;
    /** ID du dossier parent (undefined = racine) */
    parentFolderId?: string;
  };
}

export function useFolder() {
  const state = useState<FolderState>('folder-editing', () => ({
    createModal: { isOpen: false },
  }));

  // ── Rename inline ──────────────────────────────────────────────
  function requestEdit(id: string) {
    state.value = { ...state.value, editingId: id };
  }

  function clearEdit(id: string) {
    if (state.value.editingId === id) {
      state.value = { ...state.value, editingId: undefined };
    }
  }

  // ── Modale de création ─────────────────────────────────────────
  function openCreateModal(parentFolderId?: string) {
    state.value = {
      ...state.value,
      createModal: { isOpen: true, parentFolderId },
    };
  }

  function closeCreateModal() {
    state.value = {
      ...state.value,
      createModal: { isOpen: false, parentFolderId: undefined },
    };
  }

  return {
    state,
    requestEdit,
    clearEdit,
    openCreateModal,
    closeCreateModal,
  };
}
