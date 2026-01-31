import type { Folder, Note } from '~~/shared/utils/models';

/** Folder tree structure includes only folders ids*/
export interface Tree {
  [key: string]: Tree;
}

export interface Tag {
  label: string;
  count: number;
}

export type MappedTags = Map<string, Tag>;

/** Lookup is a map of all items in the store */
export type Lookup = Map<string, Note | Folder>;

/** Union of all item types in the store – derived from Lookup (single source of truth) */
export type Data = Lookup extends Map<string, infer V> ? V : never;

/** Maps ItemType → corresponding item type. Extensible: add Todo to Lookup → ItemByType<ItemType.TODO> works. */
export type ItemByType<T extends ItemType> = Extract<Data, { type: T }>;

/** StoredData is a collection of notes and folders */
export interface StoredData {
  notes: Note[];
  folders: Folder[];
}

// Props
export interface ExtractTagsProps {
  note: NotePublic;
  tags: MappedTags;
}

export interface ExtractPathProps {
  tree: Tree;
  path: string[];
}

export interface ExtractPathOptions<T extends boolean = boolean> {
  asBoolean: T;
}

// ---
interface WithFolder {
  folder: Folder;
}
interface WithContent {
  content: string;
}
interface WithTitle {
  title: string;
}

export type CreateNoteInFolderParameters = WithFolder & WithContent;
export type CreateFolderInFolderParameters = WithFolder & WithTitle;
