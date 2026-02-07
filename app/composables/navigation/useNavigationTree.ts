export function useNavigationTree() {
  const store = useDataStore();
  const { tree } = storeToRefs(store);
  const { getById } = useDataUtils();

  const navigateTree = (node: Tree): NavigationMenuItem[] => {
    const navItems: NavigationMenuItem[] = [];

    for (const id of Object.keys(node)) {
      const item = getById(id, { types: ItemType.FOLDER });

      if (!item) continue;
      const subtree = node[id] ?? {};
      const childKeys = Object.keys(subtree);

      const children = childKeys.length > 0 ? navigateTree(subtree) : undefined;

      navItems.push({
        label: item.title,
        icon: 'i-lucide-folder',
        to: item.id,
        children,
      });
    }
    return navItems;
  };

  const items = computed<NavigationMenuItem[]>(() => navigateTree(tree.value));

  return { items };
}
