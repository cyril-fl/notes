/** The notes open in the editor, oldest first: their ids only, the titles come from the notes graph. */
export const useTabs = createGlobalState(() => useLocalStorage<string[]>('tabs', []));
