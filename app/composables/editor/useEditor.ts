import type { InjectionKey, ModelRef, ShallowRef } from 'vue';
import { type Editor, useEditor } from '@tiptap/vue-3';
import { Markdown } from '@tiptap/markdown';
import StarterKit from '@tiptap/starter-kit';
import { TaskList, TaskItem } from '@tiptap/extension-list';
import Link from '@tiptap/extension-link';
import Highlight from '@tiptap/extension-highlight';
import Placeholder from '@tiptap/extension-placeholder';
import TextAlign from '@tiptap/extension-text-align';
import twitter from 'twitter-text';

interface PreviewItem {
  label: string;
  condition: boolean;
  content: string | string[] | null;
}

interface Context {
  editor: ShallowRef<Editor | undefined, Editor | undefined>;
  preview: ComputedRef<PreviewItem[]>;
}

const SYMBOL: InjectionKey<Context> = Symbol('editor-context');

export interface EditorProps {
  readonly: boolean;
  showHashtags: boolean;
  showMentions: boolean;
  showPreview: boolean;
}

export function useProvideEditorContext(
  props: EditorProps,
  content: ModelRef<string | undefined>
) {
  const isEditorReady = ref(false);

  const editor = useEditor({
    content: content.value,
    autofocus: 'end',
    contentType: 'markdown',
    extensions: [
      StarterKit.configure({
        heading: {
          levels: [1, 2, 3, 4, 5, 6],
        },
        dropcursor: {
          width: 3,
          class: 'bg-primary-500 text-primary-500 rounded-full',
        },
        link: false,
      }),
      Markdown.configure({
        indentation: { style: 'tab', size: 1 },
        markedOptions: { gfm: true },
      }),
      TaskList.configure(),
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        openOnClick: false,
        autolink: true,
        linkOnPaste: true,
        HTMLAttributes: {
          rel: 'noopener noreferrer',
          target: '_blank',
        },
        isAllowedUri: (url, ctx) => {
          const protocol = (url.split(':')[0] ?? '').toLowerCase();
          if (!['http', 'https', 'mailto'].includes(protocol)) return false;
          return ctx.defaultValidate(url);
        },
      }),
      Highlight.configure({
        multicolor: false,
      }),
      Placeholder.configure({
        placeholder: 'Commencez à écrire…',
        emptyEditorClass: 'is-editor-empty',
      }),
      TextAlign.configure({
        types: ['heading', 'paragraph'],
        alignments: ['left', 'center', 'right'],
      }),
    ],
    onUpdate: ({ editor }) => {
      if (!isEditorReady.value) return;
      const markdownContent = editor.getMarkdown() ?? '';
      handleUpdateContent(markdownContent);
    },
  });

  const hashtags = computed(() => twitter.extractHashtags(content.value ?? ''));

  const mentions = computed(() => twitter.extractMentions(content.value ?? ''));

  const preview = computed(() => [
    {
      label: 'Preview',
      condition: props.showPreview,
      content: content.value ?? '',
    },
    {
      label: 'Extracted hashtags',
      condition: props.showHashtags,
      content: hashtags.value,
    },
    {
      label: 'Extracted mentions',
      condition: props.showMentions,
      content: mentions.value,
    },
  ]);

  const handleUpdateContent = (newContent: string) => {
    if (content.value !== newContent) {
      content.value = newContent;
    }
  };

  const context: Context = {
    editor,
    preview,
  };

  provide(SYMBOL, context);

  return context;
}

export const useEditorContext = () =>
  injectContext<Context>(SYMBOL, {
    contextName: 'EditorContext',
    providerName: 'useProvideEditorContext',
  });
