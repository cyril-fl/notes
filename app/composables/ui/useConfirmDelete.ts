interface ConfirmDeleteState {
  open: boolean;
  itemId: string | null;
  itemType: 'note' | 'folder';
  resolve: ((confirmed: boolean) => void) | null;
}

const state = reactive<ConfirmDeleteState>({
  open: false,
  itemId: null,
  itemType: 'note',
  resolve: null,
});

export function useConfirmDelete() {
  function requestConfirm(
    id: string,
    type: 'note' | 'folder'
  ): Promise<boolean> {
    return new Promise((resolve) => {
      state.itemId = id;
      state.itemType = type;
      state.resolve = resolve;
      state.open = true;
    });
  }

  function confirm() {
    state.resolve?.(true);
    reset();
  }

  function cancel() {
    state.resolve?.(false);
    reset();
  }

  function reset() {
    state.open = false;
    state.itemId = null;
    state.resolve = null;
  }

  return {
    state: readonly(state),
    requestConfirm,
    confirm,
    cancel,
  };
}
