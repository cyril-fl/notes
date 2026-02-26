export const useIcons = () => {
  const icons = {
    branch: 'mingcute:git-branch-fill',
    bold: 'mingcute:bold-fill',
    code: 'mingcute:brackets-angle-fill',
    codeblock: 'mingcute:black-board-2-line',
    closelist: 'mingcute:down-fill',
    delete: 'mingcute:delete-2-fill',
    edit: 'mingcute:pencil-line',
    filter: 'mingcute:filter-2-line',
    folder: 'mingcute:folder-line',
    folderadd: 'mingcute:new-folder-line',
    folderempty: 'mingcute:folder-open-line',
    highlight: 'mingcute:mark-pen-fill',
    horizontalRule: 'i-mingcute:minimize-fill',
    italic: 'mingcute:italic-fill',
    link: 'mingcute:link-fill',
    note: 'mingcute:document-2-line',
    noteadd: 'mingcute:file-new-line',
    notebook: 'mingcute:notebook-line',
    openlist: 'mingcute:up-fill',
    orderlist: 'mingcute:list-ordered-fill',
    quote: 'mingcute:blockquote-fill',
    redo: 'mingcute:forward-2-fill',
    search: 'mingcute:search-2-line',
    strike: 'mingcute:strikethrough-fill',
    tasklist: 'mingcute:list-check-3-line',
    underline: 'mingcute:underline-fill',
    unlink: 'mingcute:unlink-2-line',
    undo: 'mingcute:back-2-fill',
    unorderedlist: 'mingcute:list-check-fill',
    warning: 'mingcute:warning-fill',
  };

  function getListStateIcon(
    isOpen: boolean,
    orientation: 'vertical' | 'horizontal' = 'vertical'
  ) {
    const openIcon = {
      vertical: {
        openlist: 'mingcute:up-fill',
        closelist: 'mingcute:down-fill',
      },
      horizontal: {
        closelist: 'mingcute:left-fill',

        openlist: 'mingcute:right-fill',
      },
    }[orientation];

    return isOpen ? openIcon.openlist : openIcon.closelist;
  }

  return { icons, getListStateIcon };
};
