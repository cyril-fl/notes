const SYMBOL: InjectionKey<SelectionContext> = Symbol('selection');

interface SelectionContext {
  selectedIds: Ref<Set<string>>;
  selectedCount: ComputedRef<number>;
  select(id: string): void;
  toggleSelect(id: string): void;
  isSelected(id: string): boolean;
  clearSelection(): void;
}

export function provideSelection() {
  const selectedIds = ref<Set<string>>(new Set());

  const selectedCount = computed(() => selectedIds.value.size);

  function select(id: string) {
    selectedIds.value = new Set([id]);
  }

  function toggleSelect(id: string) {
    const next = new Set(selectedIds.value);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedIds.value = next;
  }

  function isSelected(id: string): boolean {
    return selectedIds.value.has(id);
  }

  function clearSelection() {
    selectedIds.value = new Set();
  }

  const context: SelectionContext = {
    selectedIds,
    selectedCount,
    select,
    toggleSelect,
    isSelected,
    clearSelection,
  };

  provide(SYMBOL, context);

  return context;
}

export const useSelection = () =>
  injectContext<SelectionContext>(SYMBOL, {
    contextName: 'SelectionContext',
    providerName: 'provideSelection',
  });
