import type { Folder, Note } from '~~/shared/utils/models';

// Structure
export interface Data {
  notes: Note[];
  folders: Folder[];
}

export interface Tree {
  [key: string]: Tree;
}

export type Lookup = Map<string, Note | Folder>;

// Notes
export interface Tag {
  label: string;
  count: number;
}

export type MappedTags = Map<string, Tag>;
