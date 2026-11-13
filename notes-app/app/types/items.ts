import type { components } from './api';

type Schemas = components['schemas'];

export type Item = Schemas['domain.Item'];
export type ItemDetail = Schemas['readitem.Result'];
export type TagCount = Schemas['counttags.TagCount'];
export type TagDetail = Schemas['readtag.Result'];
export type Match = Schemas['searchnotes.Match'];
export type Trashed = Schemas['listtrash.Trashed'];
export type Neighbors = Schemas['readneighbors.Result'];
export type NotesGraph = Schemas['readgraph.Result'];
export type Recent = Schemas['listrecent.Recent'];
export type Pinned = Schemas['listpinned.Pinned'];
export type BrokenLink = Schemas['listbrokenlinks.Link'];
export type ImportResult = Schemas['importnotes.Result'];
export type FolderPath = Schemas['listfolders.Folder'];
export type User = Schemas['domain.User'];
export type Account = Schemas['users.Account'];
export type Invite = Schemas['domain.Invite'];
export type AddUserMode = Schemas['users.Mode'];
export type Snippet = Schemas['domain.Snippet'];
export type NoteFile = Schemas['domain.File'];
export type NoteVersion = Schemas['domain.NoteVersion'];
