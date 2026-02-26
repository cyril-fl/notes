import { z } from 'zod';
import { ItemType } from './models';

// TODO Refacto
export interface ServerResponse<T = null> {
  message: string;
  data: T;
}

export interface ServerErrorResponse {
  error: Error;
}

export interface RequestBody<T = null> {
  data: T;
}

export const SEARCH_TYPE_ALL = 'all' as const;

export const searchTypes = [
  SEARCH_TYPE_ALL,
  ...Object.values(ItemType),
] as const;

export type SearchType = (typeof searchTypes)[number];

export const searchQuerySchema = z.object({
  q: z.string().min(1),
  type: z.enum(searchTypes).default(SEARCH_TYPE_ALL),
  tags: z.string().optional(),
});

export type SearchQuery = z.infer<typeof searchQuerySchema>;
