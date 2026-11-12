import Image from '@tiptap/extension-image';
import { Plugin } from '@tiptap/pm/state';
import type { Editor } from '@tiptap/core';
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

/** An attachment of a note, a PDF or a recording: a link to /api/files/<id>, which the API sends as a download. */
export function isFileLink(href?: string | null) {
  return !!href?.startsWith(FILES);
}

/** In a new tab: the download leaves the note open, and an error of the API stays in that tab instead of replacing it. */
export function openFile(href: string, baseURL: string) {
  window.open(imageSrc(href, baseURL), '_blank', 'noopener');
}

/**
 * Sends the files to the API with upload and writes each in the note, an image or a link to the file. One after the
 * other, in order: a drop puts the first where it fell, the others after it.
 */
export async function insertFiles(
  editor: Editor,
  upload: (file: File) => Promise<string | null>,
  files: File[],
  pos?: number,
) {
  for (const file of files) {
    const src = await upload(file);
    // Left the note while it was sending: there is nothing left to write in.
    if (editor.isDestroyed) return;
    if (!src) continue;
    const content = file.type.startsWith('image/')
      ? { type: 'image', attrs: { src, alt: file.name } }
      : { type: 'text', text: file.name, marks: [{ type: 'link', attrs: { href: src } }] };
    if (pos === undefined) editor.chain().focus().insertContent(content).run();
    else editor.chain().focus().insertContentAt(pos, content).run();
    pos = undefined;
  }
}

/**
 * The images of a note: pasting or dropping one sends it to the API with upload, which gives what the
 * markdown cites, or null when it failed (and said so). Any other file goes the same way and becomes a link to it.
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
      const { editor } = this;
      const insert = (files: File[], pos?: number) => insertFiles(editor, upload, files, pos);

      return [
        new Plugin({
          props: {
            handlePaste(_view, event) {
              const files = [...(event.clipboardData?.files ?? [])];
              if (!files.length) return false;
              insert(files);
              return true;
            },
            handleDrop(view, event) {
              const files = [...(event.dataTransfer?.files ?? [])];
              if (!files.length) return false;
              event.preventDefault();
              insert(files, view.posAtCoords({ left: event.clientX, top: event.clientY })?.pos);
              return true;
            },
            handleClick(_view, _pos, event) {
              const href = (event.target as HTMLElement).closest('a')?.getAttribute('href');
              if (!isFileLink(href)) return false;
              openFile(href!, baseURL);
              return true;
            },
          },
        }),
      ];
    },
  });
}
