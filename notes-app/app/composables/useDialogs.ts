import { ConfirmDialog, PromptDialog, SnippetDialog } from '#components';
import type { Snippet } from '~/types/items';

/** Modals that are awaited: a text, a yes, or a snippet. */
export function useDialogs() {
  const overlay = useOverlay();
  const promptDialog = overlay.create(PromptDialog);
  const confirmDialog = overlay.create(ConfirmDialog);
  const snippetDialog = overlay.create(SnippetDialog);

  return {
    prompt: (props: { title: string; value?: string; placeholder?: string; action?: string }) =>
      promptDialog.open(props).result as Promise<string | undefined>,
    confirm: (props: { title: string; description?: string; action: string; expected?: string }) =>
      confirmDialog.open(props).result as Promise<boolean | undefined>,
    snippet: (props: { title: string; value?: Snippet }) =>
      snippetDialog.open(props).result as Promise<Snippet | undefined>,
  };
}
