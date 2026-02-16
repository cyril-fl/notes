export const useIcons = () => {
  const icons = {
    branch: 'i-lucide-git-branch',
    bold: 'i-lucide-bold',
    code: 'i-lucide-code',
    codeblock: 'i-lucide-square-code',
    delete: 'i-lucide-trash',
    edit: 'i-lucide-pencil',
    folder: 'i-lucide-folder',
    folderadd: 'i-lucide-folder-plus',
    folderempty: 'i-lucide-folder-open',
    highlight: 'i-lucide-highlighter',
    horizontalRule: 'i-lucide-minus',
    italic: 'i-lucide-italic',
    link: 'i-lucide-link',
    note: 'i-lucide-file-text',
    noteadd: 'i-lucide-file-plus',
    notebook: 'i-lucide-notebook',
    orderlist: 'i-lucide-list-ordered',
    quote: 'i-lucide-text-quote',
    redo: 'i-lucide-redo',
    strike: 'i-lucide-strikethrough',
    tasklist: 'i-lucide-check-square',
    underline: 'i-lucide-underline',
    unlink: 'i-lucide-unlink',
    undo: 'i-lucide-undo',
    unorderedlist: 'i-lucide-list',
    openlist: 'i-lucide-chevron-up',
    closelist: 'i-lucide-chevron-down',
  };

  function getListStateIcon(isOpen: boolean) {
    return isOpen ? icons.openlist : icons.closelist;
  }

  return { icons, getListStateIcon };
};
