import { ConfirmDialog, PromptDialog } from '#components';

/** Modals that are awaited: a text, or a yes. */
export function useDialogs() {
  const overlay = useOverlay();
  const promptDialog = overlay.create(PromptDialog);
  const confirmDialog = overlay.create(ConfirmDialog);

  return {
    prompt: (props: { title: string; value?: string; placeholder?: string; action?: string }) =>
      promptDialog.open(props).result as Promise<string | undefined>,
    confirm: (props: { title: string; description?: string; action: string; expected?: string }) =>
      confirmDialog.open(props).result as Promise<boolean | undefined>,
  };
}
