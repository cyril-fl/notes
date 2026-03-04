import type { NavigationTreeItem } from '~/types/ui';

export function useNavigationTree() {
  const store = useDataStore();
  const { icons } = useIcons();
  const { tree, map } = storeToRefs(store);

  const data = computed<NavigationTreeItem[]>(() => {
    const lookup = map.value;
    const treeValue = tree.value;
    const items: NavigationTreeItem[] = [];

    for (const id of Object.keys(treeValue)) {
      const item = lookup.get(id);
      if (!item || item.type !== ItemType.FOLDER) continue;

      const subtree = treeValue[id] ?? {};
      const children = makeNavigationTreeItem(subtree, lookup);

      items.push({
        id: item.id,
        label: item.title,
        icon: icons.folder,
        to: `/${item.id}/`,
        children: children.length > 0 ? children : undefined,
      });
    }

    return items;
  });

  function makeNavigationTreeItem(
    node: Tree,
    lookup: Lookup,
    visited = new Set<string>()
  ): NavigationTreeItem[] {
    const navItems: NavigationTreeItem[] = [];

    for (const id of Object.keys(node)) {
      const item = lookup.get(id);

      if (!item || item.type !== ItemType.FOLDER) continue;
      if (visited.has(id)) continue;
      visited.add(id);

      const subtree = node[id] ?? {};
      const childKeys = Object.keys(subtree);

      const children =
        childKeys.length > 0
          ? makeNavigationTreeItem(subtree, lookup, visited)
          : undefined;

      navItems.push({
        id: item.id,
        label: item.title,
        icon: icons.folder,
        to: `/${item.id}/`,
        children,
      });
    }
    return navItems;
  }

  return { folders: data };
}
