import type { Labeled, Navigable, Iconable, Clickable } from './protocol';

// Global
export type CssClass = string | string[] | Record<string, boolean>;

// Navigation types
export interface NavigationMenuItem
  extends Labeled, Partial<Navigable & Iconable> {
  children?: NavigationMenuItem[];
}

export interface NavigationMenuProps {
  item: NavigationMenuItem;
}

// Editor types
export type Level = 1 | 2 | 3 | 4 | 5 | 6;

export enum MARKS {
  BOLD = 'bold',
  ITALIC = 'italic',
  UNDERLINE = 'underline',
  STRIKE = 'strike',
  CODE = 'code',
  HIGHLIGHT = 'highlight',
}
export interface BaseEditorToolbarItem extends Iconable, Partial<Clickable> {
  kind: string;
  tooltip?: BaseEditorTooltipOptions;
}
export interface BaseEditorTooltipOptions {
  label: string;
}

export interface EditorToolbarHeadingItem extends BaseEditorToolbarItem {
  level: Level;
}

export interface EditorToolbarMarkItem extends BaseEditorToolbarItem {
  mark: MARKS;
}

export type EditorToolbarItem =
  | BaseEditorToolbarItem
  | EditorToolbarHeadingItem
  | EditorToolbarMarkItem;

//_________________________________
// Types de remplacement pour Nuxt UI
// Editor types
export interface EditorEmojiMenuItem {
  name: string;
  emoji?: string;
  fallbackImage?: string;
}

export interface EditorMentionMenuItem {
  id: string;
  label: string;

  [key: string]: unknown;
}
//_______________________
