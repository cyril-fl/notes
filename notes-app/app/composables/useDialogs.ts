import {
  ConfirmDialog,
  ConflictDialog,
  ImportDialog,
  PromptDialog,
  SavedSearchDialog,
  SnippetDialog,
} from '#components';
import type { SavedSearch, Snippet } from '~/types/items';
import type { MergePart } from '~/utils/merge';

/** Modals that are awaited: a text, a yes, a snippet, a saved search, the import done, or the text chosen after a conflict. */
export function useDialogs() {
  const overlay = useOverlay();
  const promptDialog = overlay.create(PromptDialog);
  const confirmDialog = overlay.create(ConfirmDialog);
  const snippetDialog = overlay.create(SnippetDialog);
  const savedSearchDialog = overlay.create(SavedSearchDialog);
  const importDialog = overlay.create(ImportDialog);
  const conflictDialog = overlay.create(ConflictDialog);

  return {
    prompt: (props: {
      title: string;
      value?: string;
      placeholder?: string;
      action?: string;
      multiline?: boolean;
      alternative?: { label: string; icon?: string; onSelect: () => void };
    }) => promptDialog.open(props).result as Promise<string | undefined>,
    confirm: (props: { title: string; description?: string; action: string; expected?: string }) =>
      confirmDialog.open(props).result as Promise<boolean | undefined>,
    snippet: (props: { title: string; value?: Snippet }) =>
      snippetDialog.open(props).result as Promise<Snippet | undefined>,
    savedSearch: (props: { title: string; value?: SavedSearch }) =>
      savedSearchDialog.open(props).result as Promise<SavedSearch | undefined>,
    importNotes: () => importDialog.open().result as Promise<void>,
    resolveConflict: (parts: MergePart[]) => conflictDialog.open({ parts }).result as Promise<string>,
  };
}
