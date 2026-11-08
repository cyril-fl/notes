import type { EditorHandler, EditorSuggestionMenuItem, EditorSuggestionMenuProps, EditorToolbarItem } from '@nuxt/ui';
import type { Editor, Extensions } from '@tiptap/vue-3';
import type { MaybeRefOrGetter } from 'vue';
import type { Feature } from './features';

/** A floating toolbar, shown over the text when shouldShow says so. */
export interface EditorBubble {
  pluginKey: string;
  items: EditorToolbarItem[][];
  shouldShow: (props: { editor: Pick<Editor, 'isActive' | 'getAttributes'> }) => boolean;
}

/** A menu opened by typing its char; items can follow the data they come from. */
export type EditorMenu = Omit<EditorSuggestionMenuProps, 'editor' | 'items'> & {
  pluginKey: string;
  items: MaybeRefOrGetter<EditorSuggestionMenuItem[]>;
};

/** What a feature of the note editor adds to it: the editor merges them, a new feature does not touch NoteEditor.vue. */
export interface EditorModule {
  /** Left out when the account lacks it, as usePro().has says. */
  feature?: Feature;
  extensions?: Extensions;
  starterKit?: object;
  handlers?: Record<string, EditorHandler>;
  toolbar?: EditorToolbarItem[][];
  bubbles?: EditorBubble[];
  menus?: EditorMenu[];
  onClick?: (event: MouseEvent) => void;
}

/** A handler that can always run, inactive unless told otherwise. */
export function command(
  execute: EditorHandler['execute'],
  isActive: EditorHandler['isActive'] = () => false,
): EditorHandler {
  return { canExecute: () => true, execute, isActive };
}

/** The modules the account has, in order: lists are joined, handlers and starter kit options merged, every click handler called. */
export function mergeModules(all: EditorModule[], has: (feature: Feature) => boolean) {
  const modules = all.filter((module) => !module.feature || has(module.feature));
  return {
    extensions: modules.flatMap((module) => module.extensions ?? []),
    starterKit: Object.assign({}, ...modules.map((module) => module.starterKit)),
    handlers: Object.assign({}, ...modules.map((module) => module.handlers)) as Record<string, EditorHandler>,
    toolbar: modules.flatMap((module) => module.toolbar ?? []),
    bubbles: modules.flatMap((module) => module.bubbles ?? []),
    menus: modules.flatMap((module) => module.menus ?? []),
    onClick: (event: MouseEvent) => modules.forEach((module) => module.onClick?.(event)),
  };
}
