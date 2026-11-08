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
      imageNormal: sizeHandler(null),
      imageSmall: sizeHandler('small'),
      imageMedium: sizeHandler('medium'),
      imageLarge: sizeHandler('large'),
    },
    // Fixed sizes rather than handles: they work with a finger too. One icon, drawn at the size it gives, after
    // the image's own size.
    bubbles: [
      {
        pluginKey: 'imageMenu',
        items: [
          [{ kind: 'imageNormal', icon: icons.imageNormal, tooltip: { text: t('note.imageNormal') } }],
          [
            {
              kind: 'imageSmall',
              icon: icons.image,
              ui: { leadingIcon: 'size-3' },
              tooltip: { text: t('note.imageSmall') },
            },
            {
              kind: 'imageMedium',
              icon: icons.image,
              ui: { leadingIcon: 'size-4' },
              tooltip: { text: t('note.imageMedium') },
            },
            {
              kind: 'imageLarge',
              icon: icons.image,
              ui: { leadingIcon: 'size-5' },
              tooltip: { text: t('note.imageLarge') },
            },
          ],
        ],
        shouldShow: ({ editor }) => editor.isActive('image'),
      },
    ],
  };
}
