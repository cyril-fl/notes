export enum NAVIGATION {
  home = '/',
  graph = '/diagram/',
  notes = '/editor/',
  newNote = '/editor/new/',
}

export function useNavigation() {
  // Define

  // Data
  const data: NavigationMenuItem[] = [
    {
      id: 'main',
      label: 'menu.navigation.title',
      children: [
        {
          id: 'home',
          label: 'pages.home.title',
          icon: 'i-lucide-home',
          to: NAVIGATION.home,
        },
        {
          id: 'graph',
          label: 'pages.graph.title',
          icon: 'i-lucide-git-branch',
          to: NAVIGATION.graph,
        },
        {
          id: 'notes',
          label: 'pages.notes.title',
          icon: 'i-lucide-list',
          to: NAVIGATION.notes,
        },
        {
          id: 'new-note',
          label: 'pages.notes.new.title',
          icon: 'i-lucide-plus-circle',
          to: NAVIGATION.newNote,
        },
      ],
    },
  ];

  return {
    links: data,
  };
}
