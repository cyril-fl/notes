// The settled layout of the whole graph, per account, in this browser (IndexedDB, read by the layout worker too).
// A cache: missing or unreadable, the layout is computed again.
const DB = 'notes-graph';
const STORE = 'layouts';

export type SavedLayout = { signature: string; positions: Map<string, { x: number; y: number }> };

function request<T>(req: IDBRequest<T>) {
  return new Promise<T>((resolve, reject) => {
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

async function store(mode: IDBTransactionMode) {
  const open = indexedDB.open(DB, 1);
  open.onupgradeneeded = () => open.result.createObjectStore(STORE);
  const db = await request(open);
  const transaction = db.transaction(STORE, mode);
  transaction.oncomplete = () => db.close(); // An open connection would hold forgetLayouts back.
  return transaction.objectStore(STORE);
}

export async function readLayout(key: string) {
  return request<SavedLayout | undefined>((await store('readonly')).get(key));
}

export async function saveLayout(key: string, layout: SavedLayout) {
  await request((await store('readwrite')).put(layout, key));
}

/** Forgets every saved layout: after a logout, the next account must not find this one's notes. */
export function forgetLayouts() {
  indexedDB.deleteDatabase(DB);
}
