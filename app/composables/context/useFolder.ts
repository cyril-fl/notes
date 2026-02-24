import { useActions } from '#imports';

const SYMBOL = Symbol('folder');

interface Context {
  current: Ref<string | null>;
  editingId: Ref<string | null>;
  actions: ReturnType<typeof useActions>;
  reset(): void;
  set(folderId: string): void;
  requestEdit(id: string): void;
  clearEdit(): void;
}

export function provideCurrentFolder() {
  const current = ref<string | null>(null);
  const editingId = ref<string | null>(null);

  function requestEdit(id: string) {
    editingId.value = id;
  }

  function clearEdit() {
    editingId.value = null;
  }

  const actions = useActions(current, { requestEdit });

  function reset() {
    current.value = null;
  }

  function set(folderId: string) {
    current.value = folderId;
  }

  provide(SYMBOL, {
    current,
    editingId,
    actions,
    reset,
    set,
    requestEdit,
    clearEdit,
  } satisfies Context);
}

export const useCurrentFolder = () =>
  injectContext<Context>(SYMBOL, {
    contextName: 'FolderContext',
    providerName: 'provideCurrentFolder',
  });
