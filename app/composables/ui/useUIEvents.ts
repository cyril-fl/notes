type Listener<T extends unknown[]> = (...args: T) => void;

function createEventBus<T extends unknown[]>() {
  const listeners = new Set<Listener<T>>();

  function on(cb: Listener<T>): () => void {
    listeners.add(cb);
    return () => listeners.delete(cb);
  }

  function emit(...args: T): void {
    listeners.forEach((cb) => cb(...args));
  }

  return { on, emit };
}

const folderCardUpdate = createEventBus<[id: string]>();
const folderNavOpen = createEventBus<[id: string]>();

export function useUIEvents() {
  return {
    onFolderCardUpdate: folderCardUpdate.on,
    emitFolderCardUpdate: folderCardUpdate.emit,
    onFolderNavOpen: folderNavOpen.on,
    emitFolderNavOpen: folderNavOpen.emit,
  };
}
