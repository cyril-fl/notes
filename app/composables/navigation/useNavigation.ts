export enum NAVIGATION {
  home = '/',
  graph = '/graph/',
  notes = '/editor/',
  newNote = '/editor/new/',
}

export function useNavigation() {
  const { icons } = useIcons();

  const data: NavigationMenuItem[] = [
    {
      id: 'main',
      label: 'menu.navigation.title',
      children: [
        {
          id: 'graph',
          label: 'pages.graph.title',
          icon: icons.branch,
          to: NAVIGATION.graph,
        },
        {
          id: 'notes',
          label: 'pages.editor.title',
          icon: icons.unorderedlist,
          to: NAVIGATION.notes,
        },
        {
          id: 'new-note',
          label: 'pages.editor.new.title',
          icon: icons.noteadd,
          to: NAVIGATION.newNote,
        },
      ],
    },
  ];

  return {
    links: data,
  };
}
