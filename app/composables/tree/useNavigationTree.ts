import type { NavigationTreeItem } from '~/types/ui';

export function useNavigationTree() {
  const store = useDataStore();
  const { icons } = useIcons();
  const { tree, data: _data } = storeToRefs(store);
  const { getById } = useDataUtils();

  const data = computed<NavigationTreeItem[]>(() => [
    {
      id: 'folders',
      label: 'menu.navigation.folders.title',
      children: makeNavigationTreeItem(tree.value),
    },
  ]);

  function makeNavigationTreeItem(node: Tree): NavigationTreeItem[] {
    const navItems: NavigationTreeItem[] = [];

    for (const id of Object.keys(node)) {
      const item = getById(id, { types: ItemType.FOLDER });

      if (!item) continue;
      const subtree = node[id] ?? {};
      const childKeys = Object.keys(subtree);

      const children =
        childKeys.length > 0 ? makeNavigationTreeItem(subtree) : undefined;

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
