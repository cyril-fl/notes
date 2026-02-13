import type { NavigationTreeItem } from '~/types/ui';

export function useNavigationTree() {
  const store = useDataStore();
  const { tree } = storeToRefs(store);
  const { getById } = useDataUtils();

  const items = computed<NavigationTreeItem[]>(() =>
    makeNavigationTreeItem(tree.value)
  );

  const data = computed<NavigationTreeItem[]>(() => [
    {
      label: 'menu.navigation.folders.title',
      children: items.value,
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
        label: item.title,
        icon: 'i-lucide-folder',
        to: `/${item.id}/`,
        children,
      });
    }
    return navItems;
  }

  return { folders: data };
}
