import Image from '@tiptap/extension-image';
import { Plugin } from '@tiptap/pm/state';
import { mergeAttributes, type JSONContent } from '@tiptap/vue-3';

const FILES = '/api/files/';

export type ImageSize = 'small' | 'medium' | 'large';

/** The size of an image is the title of its markdown, like the cards of links: ![alt](src "small"). Left out, the image keeps its own width. */
export function imageSize(title?: string | null): ImageSize | null {
  const size = title?.toLowerCase();
  return size === 'small' || size === 'medium' || size === 'large' ? size : null;
}

/** What the markdown cites, /api/files/<id>, is on the API: in development it runs on another port. */
export function imageSrc(src: string, baseURL: string) {
  return src.startsWith(FILES) ? baseURL + src : src;
}

/** An image copied from a note comes back with the API's address: the markdown keeps /api/files/<id>. */
export function citedSrc(src: string, baseURL: string) {
  return src.startsWith(baseURL + FILES) ? src.slice(baseURL.length) : src;
}

function images(files?: FileList | null) {
  return [...(files ?? [])].filter((file) => file.type.startsWith('image/'));
}

/**
 * The images of a note: pasting or dropping one sends it to the API with upload, which gives what the
 * markdown cites, or null when it failed (and said so).
 */
export function noteImage(baseURL: string, upload: (file: File) => Promise<string | null>) {
  return Image.extend({
    addAttributes() {
      return {
        ...this.parent?.(),
        src: { default: null, parseHTML: (element) => citedSrc(element.getAttribute('src') ?? '', baseURL) || null },
        size: {
          default: null,
          parseHTML: (element) => imageSize(element.dataset.size),
          renderHTML: ({ size }) => (size ? { 'data-size': size, class: `note-image-${size}` } : {}),
        },
      };
    },

    parseMarkdown(token, helpers) {
      const node = this.parent!(token, helpers) as JSONContent;
      const size = imageSize(node.attrs?.title);
      return size ? { ...node, attrs: { ...node.attrs, title: null, size } } : node;
    },

    renderMarkdown(node, helpers, context) {
      const size = node.attrs?.size;
      return this.parent!(size ? { ...node, attrs: { ...node.attrs, title: size } } : node, helpers, context) as string;
    },

    renderHTML({ HTMLAttributes }) {
      return [
        'img',
        mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, {
          src: imageSrc(HTMLAttributes.src ?? '', baseURL),
        }),
      ];
    },

    addProseMirrorPlugins() {
      const { editor, name } = this;
      // One after the other, in order: a drop puts the first where it fell, the others after it.
      async function insert(files: File[], pos?: number) {
        for (const file of files) {
          const src = await upload(file);
          if (!src) continue;
          const content = { type: name, attrs: { src, alt: file.name } };
          if (pos === undefined) editor.chain().focus().insertContent(content).run();
          else editor.chain().focus().insertContentAt(pos, content).run();
          pos = undefined;
        }
      }

      return [
        new Plugin({
          props: {
            handlePaste(_view, event) {
              const files = images(event.clipboardData?.files);
              if (!files.length) return false;
              insert(files);
              return true;
            },
            handleDrop(view, event) {
              const files = images(event.dataTransfer?.files);
              if (!files.length) return false;
              event.preventDefault();
              insert(files, view.posAtCoords({ left: event.clientX, top: event.clientY })?.pos);
              return true;
            },
          },
        }),
      ];
    },
  });
}
