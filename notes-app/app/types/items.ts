import type { components } from './api';

type Schemas = components['schemas'];

export type Item = Schemas['domain.Item'];
export type ItemDetail = Schemas['readitem.Result'];
export type TagCount = Schemas['counttags.TagCount'];
export type TagDetail = Schemas['readtag.Result'];
export type Match = Schemas['searchnotes.Match'];
export type Trashed = Schemas['listtrash.Trashed'];
export type Neighbors = Schemas['readneighbors.Result'];
