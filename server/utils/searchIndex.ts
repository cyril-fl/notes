import MiniSearch, { type SearchResult } from 'minisearch';
import type { DataSchema } from '../../shared/types/models';
import { ItemType, dataParams } from '../../shared/types/models';
import { SEARCH_TYPE_ALL, type SearchType } from '../../shared/types/api';

interface SearchDocument {
  id: string;
  type: string;
  title: string;
  content: string;
}

interface SearchFilters {
  type?: SearchType;
  tags?: string[];
}

let _miniSearch: MiniSearch<SearchDocument> | null = null;
const _documents: Map<string, DataSchema> = new Map();

function extractTitle(content: string): string {
  for (let level = 1; level <= 6; level++) {
    const pattern = new RegExp(`^${'#'.repeat(level)}\\s+(.+)$`, 'm');
    const match = content.match(pattern);
    if (match?.[1]) return match[1].trim();
  }
  return '';
}

function extractHashtags(content: string): string[] {
  const matches = content.match(/#(\w+)/g);
  return matches ? matches.map((m) => m.slice(1)) : [];
}

function toSearchDocument(doc: DataSchema): SearchDocument {
  return {
    id: doc.id,
    type: doc.type,
    title: doc.type === ItemType.NOTE ? extractTitle(doc.content) : doc.title,
    content: doc.type === ItemType.NOTE ? doc.content : '',
  };
}

function createMiniSearch(): MiniSearch<SearchDocument> {
  return new MiniSearch<SearchDocument>({
    fields: ['title', 'content'],
    storeFields: ['id', 'type'],
    searchOptions: {
      fuzzy: 0.2,
      prefix: true,
      boost: { title: 2 },
    },
  });
}

async function initIndex(): Promise<void> {
  if (_miniSearch) return;

  const collection = await getMongoCollection();
  const rawDocs = await collection.find({}).toArray();

  _miniSearch = createMiniSearch();
  _documents.clear();

  for (const raw of rawDocs) {
    const parsed = dataParams.safeParse(raw.value);
    if (!parsed.success) continue;

    const doc = parsed.data;
    _documents.set(doc.id, doc);
    _miniSearch.add(toSearchDocument(doc));
  }
}

export async function search(
  query: string,
  filters: SearchFilters = {}
): Promise<DataSchema[]> {
  await initIndex();

  const results: SearchResult[] = _miniSearch!.search(query);

  const filtered = results.filter((result) => {
    if (filters.type && filters.type !== SEARCH_TYPE_ALL) {
      if (result.type !== filters.type) return false;
    }

    if (filters.tags?.length) {
      const doc = _documents.get(result.id);
      if (!doc || doc.type !== ItemType.NOTE) return false;
      const hashtags = extractHashtags(doc.content);
      if (!filters.tags.every((tag) => hashtags.includes(tag))) return false;
    }

    return true;
  });

  return filtered
    .slice(0, 50)
    .map((result) => _documents.get(result.id))
    .filter((doc): doc is DataSchema => !!doc);
}

export async function addToIndex(doc: DataSchema): Promise<void> {
  await initIndex();
  _documents.set(doc.id, doc);
  _miniSearch!.add(toSearchDocument(doc));
}

export async function removeFromIndex(id: string): Promise<void> {
  await initIndex();
  const doc = _documents.get(id);
  if (!doc) return;
  _miniSearch!.remove(toSearchDocument(doc));
  _documents.delete(id);
}

export async function updateInIndex(doc: DataSchema): Promise<void> {
  await removeFromIndex(doc.id);
  await addToIndex(doc);
}
