/** Every icon of the app, by what it stands for: the one place to change an icon everywhere. */
export function useIcons() {
  const icons = {
    // Items
    note: 'i-lucide-file-text',
    noteAdd: 'i-lucide-square-pen',
    noteNotFound: 'i-lucide-file-x',
    notes: 'i-lucide-notebook-text',
    folder: 'i-lucide-folder',
    folderAdd: 'i-lucide-folder-plus',
    folderEmpty: 'i-lucide-folder-open',
    folderNotFound: 'i-lucide-folder-x',
    tag: 'i-lucide-hash',
    recent: 'i-lucide-clock',
    pin: 'i-lucide-pin',
    unpin: 'i-lucide-pin-off',
    favorite: 'i-lucide-star',
    unfavorite: 'i-lucide-star-off',
    rename: 'i-lucide-pencil',
    delete: 'i-lucide-trash-2',
    trash: 'i-lucide-trash-2',

    // Navigation
    search: 'i-lucide-search',
    searchEmpty: 'i-lucide-search-x',
    graph: 'i-lucide-waypoints',
    settings: 'i-lucide-settings',
    home: 'i-lucide-arrow-left',
    back: 'i-lucide-arrow-big-left',
    forward: 'i-lucide-arrow-big-right',
    chevronRight: 'i-lucide-chevron-right',
    chevronUp: 'i-lucide-chevron-up',
    chevronDown: 'i-lucide-chevron-down',
    close: 'i-lucide-x',

    // Views and settings
    list: 'i-lucide-list',
    grid: 'i-lucide-layout-grid',
    lightMode: 'i-lucide-sun',
    darkMode: 'i-lucide-moon',
    export: 'i-lucide-download',
    import: 'i-lucide-upload',

    // Graph
    zoomIn: 'i-lucide-plus',
    zoomOut: 'i-lucide-minus',
    recenter: 'i-lucide-locate-fixed',

    // Editor
    heading1: 'i-lucide-heading-1',
    heading2: 'i-lucide-heading-2',
    heading3: 'i-lucide-heading-3',
    bold: 'i-lucide-bold',
    italic: 'i-lucide-italic',
    strike: 'i-lucide-strikethrough',
    code: 'i-lucide-code',
    link: 'i-lucide-link',
    orderedList: 'i-lucide-list-ordered',
    taskList: 'i-lucide-list-checks',
    quote: 'i-lucide-quote',
  };

  return { icons };
}
