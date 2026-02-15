interface FolderState {
  isEditing: boolean;
  targetId?: string;
}

export function useFolder() {
  const isEditing = useState<FolderState>('folder-editing', () => ({
    isEditing: false,
  }));

  function setState(state: Partial<FolderState>) {
    isEditing.value = { ...isEditing.value, ...state };
  }

  return {
    isEditing,
    setState,
  };
}
