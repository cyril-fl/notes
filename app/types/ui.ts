// Global
export type CssClass = string | string[] | Record<string, boolean>;

// Protocol
interface Labeled {
  label: string;
}

interface Navigable {
  to: string;
}

export interface Iconable {
  icon: string;
}

// Button types
export interface ButtonProps extends Partial<Labeled & Navigable & Iconable> {
  icon?: string;

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

export type NavigationGroupProps = Labeled;

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

export interface EditorToolbarItemTooltip {
  text: string;
}

export interface EditorToolbarItem {
  kind: string;
  icon?: string;
  tooltip?: EditorToolbarItemTooltip;
  mark?: string;
  level?: number;
  align?: string;
}

// Navigation types

export interface TreeItem {
  label: string;
  icon?: string;
  children?: TreeItem[];
  defaultExpanded?: boolean;
  slot?: string;

  [key: string]: unknown;
}

//_______________________
