/** The sidebar lists shown above the tree, one choice per device: kept across reloads, like the view. */
export const useSidebarSections = createGlobalState(() =>
  useLocalStorage('sidebar-sections', { favorites: true, pinned: true, recent: true }, { mergeDefaults: true }),
);
