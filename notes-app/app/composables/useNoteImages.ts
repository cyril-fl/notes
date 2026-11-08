import type { ImageSize } from '~/utils/noteImage';

/** The images of the note editor, pasted or dropped, and their size picked over the selected one. */
export function useNoteImages(baseURL: string, upload: (file: File) => Promise<string | null>): EditorModule {
  const { icons } = useIcons();
  const { t } = useI18n();

  function sizeHandler(size: ImageSize | null) {
    return command(
      (editor) => editor.chain().focus().updateAttributes('image', { size }),
      (editor) => editor.isActive('image') && (editor.getAttributes('image').size ?? null) === size,
    );
  }

  return {
    extensions: [noteImage(baseURL, upload)],
    handlers: {
      imageSmall: sizeHandler('small'),
      imageMedium: sizeHandler('medium'),
      imageLarge: sizeHandler(null),
    },
    // Fixed sizes rather than handles: they work with a finger too, like the cards of links.
    bubbles: [
      {
        pluginKey: 'imageMenu',
        items: [
          [
            { kind: 'imageSmall', icon: icons.linkMini, tooltip: { text: t('note.imageSmall') } },
            { kind: 'imageMedium', icon: icons.imageMedium, tooltip: { text: t('note.imageMedium') } },
            { kind: 'imageLarge', icon: icons.linkLarge, tooltip: { text: t('note.imageLarge') } },
          ],
        ],
        shouldShow: ({ editor }) => editor.isActive('image'),
      },
    ],
  };
}
