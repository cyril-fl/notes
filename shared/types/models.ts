import { z } from 'zod';
import type { Item, Folder, Note } from '../utils/models';

export enum ItemType {
  FOLDER = 'folder',
  NOTE = 'note',
}

/** Zod schema for {@link Item} class constructor props */
const serverParams = z.object({
  id: z.string(),
  createdAt: z.coerce.date(), // Accepte string ou Date, convertit en Date
  updatedAt: z.coerce.date(), // Accepte string ou Date, convertit en Date
});

const baseItemParams = serverParams.extend({
  path: z.array(z.string()),
});

/** Zod schema for {@link Note} class constructor props */
export const noteParams = baseItemParams.extend({
  type: z.literal(ItemType.NOTE),
  content: z.string(),
});

/** Zod schema for {@link Folder} class constructor props */
export const folderParams = baseItemParams.extend({
  type: z.literal(ItemType.FOLDER),
  title: z.string(),
  childrenIds: z.array(z.string()),
});

/** Union type for data returned from API, used in API responses (includes server-generated fields) */
export const dataParams = z.discriminatedUnion('type', [
  folderParams,
  noteParams,
]);

const serverGeneratedFields = {
  id: true,
  createdAt: true,
  updatedAt: true,
} as const;

export const folderDraftParams = folderParams.omit(serverGeneratedFields);
export const noteDraftParams = noteParams.omit(serverGeneratedFields);

/** Union type for data sent to API, used in API requests (create) */
export const dataDraft = z.discriminatedUnion('type', [
  folderDraftParams,
  noteDraftParams,
]);
/** Union type for data sent to API, used in API requests (update) */
export const dataDraftPartial = z
  .object({
    path: z.array(z.string()),
    content: z.string(),
    title: z.string(),
    childrenIds: z.array(z.string()),
  })
  .partial();
// Schema types = data received from API (server → client), includes all DB fields
/** Type of {@link dataParams} */
export type ItemSchema = z.infer<typeof dataParams>;
/** Data returned from API for {@link Folder}, includes all DB fields (id, createdAt, updatedAt) */
export type FolderSchema = z.infer<typeof folderParams>;
/** Data returned from API for {@link Note}, includes all DB fields (id, createdAt, updatedAt) */
export type NoteSchema = z.infer<typeof noteParams>;
/** Union type for data returned from API, used in API responses */
export type DataSchema = z.infer<typeof dataParams>;

// Draft types = data sent to API (client → server), excludes server-generated fields (id, createdAt, updatedAt)
/** Data sent to API to create/update a {@link Folder}, excludes server-generated fields */
export type DraftFolder = z.infer<typeof folderDraftParams>;
/** Data sent to API to create/update a {@link Note}, excludes server-generated fields */
export type DraftNote = z.infer<typeof noteDraftParams>;
/** Union type for data sent to API, used in API requests (create/update) */
export type DraftData = z.infer<typeof dataDraft>;

// Public types = only public properties of models
/** Utility type to extract only public properties of a Class */
type Public<T> = Pick<T, keyof T>;

/** Same as {@link Item} class */
export type ItemPublic = Public<Item>;
/** Same as {@link Folder} class */
export type FolderPublic = Public<Folder>;
/** Same as {@link Note} class */
export type NotePublic = Public<Note>;
/** Union of public types {@link FolderPublic} and {@link NotePublic} */
export type DataPublic = FolderPublic | NotePublic;

export type Data = Folder | Note;
