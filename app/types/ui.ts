import type { Labeled, Navigable, Iconable, Clickable } from './protocol';

// Global
export type CssClass = string | string[] | Record<string, boolean>;

// Button types
export interface ButtonProps extends Partial<Labeled & Navigable & Iconable> {
  [key: string]: unknown;
}

// Navigation types
export interface NavigationMenuItem
  extends Labeled, Partial<Navigable & Iconable> {
  children?: NavigationMenuItem[];
}

export interface NavigationMenuProps {
  item: NavigationMenuItem;
}

// Editor types
export enum MARKS {
  BOLD = 'bold',
  ITALIC = 'italic',
  UNDERLINE = 'underline',
  STRIKE = 'strike',
  CODE = 'code',
}

export enum ALIGNMENTS {
  LEFT = 'left',
  CENTER = 'center',
  RIGHT = 'right',
}

export interface BaseEditorToolbarItem extends Iconable, Partial<Clickable> {
  kind: string;
  tooltip?: BaseEditorTooltipOptions;
}
export interface BaseEditorTooltipOptions {
  label: string;
}

export interface EditorToolbarHeadingItem extends BaseEditorToolbarItem {
  level: 1 | 2 | 3 | 4 | 5 | 6;
}

export interface EditorToolbarMarkItem extends BaseEditorToolbarItem {
  mark: MARKS;
}

export interface EditorToolbarAlignItem extends BaseEditorToolbarItem {
  align: ALIGNMENTS;
}

export type EditorToolbarItem =
  | BaseEditorToolbarItem
  | EditorToolbarHeadingItem
  | EditorToolbarMarkItem
  | EditorToolbarAlignItem;

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
