import type { EditorEmojiMenuItem } from '~/types/ui';
import { gitHubEmojis } from '@tiptap/extension-emoji';

export const useEditorEmoji = (): EditorEmojiMenuItem[] => {
  const emojiItems: EditorEmojiMenuItem[] = gitHubEmojis.filter(
    (emoji) => !emoji.name.startsWith('regional_indicator_')
  );

  return emojiItems;
};
