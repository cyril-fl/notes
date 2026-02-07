import type { InjectionKey, ShallowRef } from 'vue';
import { type Editor, useEditor } from '@tiptap/vue-3';
import { Markdown } from '@tiptap/markdown';
import StarterKit from '@tiptap/starter-kit';
import Emoji from '@tiptap/extension-emoji';
import striptags from 'striptags';
import twitter from 'twitter-text';

interface PreviewItem {
  label: string;
  condition: boolean;
  content: string | string[] | null;
}

interface Context {
  editor: ShallowRef<Editor | undefined, Editor | undefined>;
  content: Ref<string | null>;
  preview: ComputedRef<PreviewItem[]>;
}

const SYMBOL: InjectionKey<Context> = Symbol('editor-context');

export interface EditorProps {
  readonly: boolean;
  showHashtags: boolean;
  showMentions: boolean;
  showPreview: boolean;
}

export function useProvideEditorContext(props: EditorProps) {
  const content = ref<string | null>(null);

  const editor = useEditor({
    autofocus: 'end',
    contentType: 'markdown',
    extensions: [
      Markdown.configure({
        indentation: { style: 'tab', size: 1 },
      }),
      StarterKit.configure({
        heading: {
          levels: [1, 2, 3],
        },
      }),
      Emoji.configure({
        enableEmoticons: true,
      }),
    ],
    onUpdate: ({ editor }) => {
      const json = editor.getJSON();
      const markdownContent = editor.markdown?.serialize(json) ?? null;

      handleUpdateContent(markdownContent);
    },
  });

  const hashtags = computed(() => twitter.extractHashtags(content.value ?? ''));

  const mentions = computed(() => twitter.extractMentions(content.value ?? ''));

  const preview = computed(() => [
    {
      label: 'Preview',
      condition: props.showPreview,
      content: content.value ?? '  ',
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

  const handleUpdateContent = (newContent: string | null) => {
    if (newContent === null) {
      content.value = null;
      return;
    }
    const normalized = striptags(newContent)
      .replace(/&nbsp;/gi, ' ')
      .replace(/\u00A0/g, ' ')
      .trim();
    content.value = normalized.length !== 0 ? normalized : null;
  };

  const context: Context = {
    editor,
    content,
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
