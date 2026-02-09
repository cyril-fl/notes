import { useNavigationTree } from '~/composables/navigation/useNavigationTree';

export function useNavigation() {
  // Define
  enum PATHS {
    home = '/',
    graph = '/diagram/',
    notes = '/notes/',
    newNote = '/notes/new/',
  }

  const { items } = useNavigationTree();

  // Data
  const data: NavigationMenuItem[] = [
    {
      label: 'menu.navigation.title',
      children: [
        {
          label: 'pages.home.title',
          icon: 'i-lucide-home',
          to: PATHS.home,
        },
        {
          label: 'pages.graph.title',
          icon: 'i-lucide-git-branch',
          to: PATHS.graph,
        },
        {
          label: 'pages.notes.title',
          icon: 'i-lucide-list',
          to: PATHS.notes,
        },
        {
          label: 'pages.notes.new.title',
          icon: 'i-lucide-plus-circle',
          to: PATHS.newNote,
        },
      ],
    },
    {
      label: 'menu.navigation.folders.title',
      children: items.value,
    },
  ];

  // Methods
  // const isCurrent = (path: string) => route.path === path;

  return {
    links: data,
  };
}
