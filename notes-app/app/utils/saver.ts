export type Changes = { title?: string; content?: string };

// send returns false when the changes cannot go yet: they wait for the next ones.
export type Send = (changes: Changes) => Promise<boolean | void>;

/** Sends one request at a time and merges what comes meanwhile: two quick edits of a new note must not create two notes. */
export function createSaver(send: Send, delay = 1500) {
  let pending: Changes | null = null;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let running: Promise<void> | null = null;

  function keep(changes: Changes) {
    pending = { ...changes, ...pending };
  }

  async function drain() {
    while (pending) {
      const changes: Changes = pending;
      pending = null;
      try {
        if ((await send(changes)) === false) return keep(changes);
      } catch {
        return keep(changes); // Sent again with the next edit, or when leaving the page.
      }
    }
  }

  function flush(): Promise<void> {
    clearTimeout(timer);
    running ??= drain().finally(() => {
      running = null;
    });
    return running;
  }

  function push(changes: Changes) {
    pending = { ...pending, ...changes };
    clearTimeout(timer);
    timer = setTimeout(flush, delay);
  }

  /** Empties the queue and gives what was in it. */
  function take(): Changes | null {
    const changes = pending;
    pending = null;
    clearTimeout(timer);
    return changes;
  }

  return { push, flush, take, isBusy: () => running !== null };
}
