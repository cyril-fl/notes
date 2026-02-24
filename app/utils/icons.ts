export const useIcons = () => {
  const icons = {
    branch: 'mingcute:git-branch-fill',
    bold: 'mingcute:bold-fill',
    code: 'mingcute:brackets-angle-fill',
    codeblock: 'mingcute:black-board-2-line',
    delete: 'mingcute:delete-2-fill',
    edit: 'mingcute:pencil-line',
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
    orderlist: 'mingcute:list-ordered-fill',
    quote: 'mingcute:blockquote-fill',
    redo: 'mingcute:forward-2-fill',
    strike: 'mingcute:strikethrough-fill',
    tasklist: 'mingcute:list-check-3-line',
    underline: 'mingcute:underline-fill',
    unlink: 'mingcute:unlink-2-line',
    undo: 'mingcute:back-2-fill',
    unorderedlist: 'mingcute:list-check-fill',
    openlist: 'mingcute:up-fill',
    closelist: 'mingcute:down-fill',
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
