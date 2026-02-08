// Protocol

export interface Labeled {
  label: string;
}

export interface Navigable {
  to: string;
}

export interface Iconable {
  icon: string;
}

export interface Clickable {
  onClick: () => void;
}
