// One choice for every view, kept across reloads.
export const useView = createGlobalState(() => useLocalStorage<'list' | 'grid'>('view', 'list'));
