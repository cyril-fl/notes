import { createDefu } from 'defu';

/** Merge arrays by concatenating and deduplicating (any array key). */
export const defuDedupArrays = createDefu((obj, key, value) => {
  if (!Array.isArray(value)) return false;
  const o = obj as Record<string, unknown>;
  const k = key as string;
  const existing = Array.isArray(o[k]) ? (o[k] as unknown[]) : [];
  o[k] = [...new Set([...existing, ...value])];
  return true;
});

export const extractTags = ({ note, tags }: ExtractTagsProps) => {
  for (const label of note.hashtags) {
    const tag = tags.get(label);

    if (tag) {
      tag.count++;
      continue;
    }

    tags.set(label, {
      label,
      count: 1,
    });
  }

  return tags;
};

export const rebuildTree = (folders: Folder[]) => {
  const tree: Tree = {};

  for (const folder of folders) {
    const raw = folder.path.filter((s) => s !== '.');
    let current: Tree = tree;
    for (const segment of raw) {
      current = current[segment] ??= {};
    }
  }

  return tree;
};

// Tree
export function exploreTree({
  tree,
  path,
  options,
}: Omit<ExtractPathProps, 'options'> & {
  options: { asBoolean: true };
}): boolean;
export function exploreTree(
  params: ExtractPathProps & { options?: { asBoolean?: false } }
): string[] | undefined;
export function exploreTree({
  tree,
  path,
  options,
}: ExtractPathProps & { options?: Partial<ExtractPathOptions> }):
  | boolean
  | string[]
  | undefined {
  const returnBool = !!options?.asBoolean;

  let current: Tree = tree;
  for (const seg of path) {
    const next = current[seg];
    if (!next) return returnBool ? false : undefined;
    current = next;
  }
  return returnBool ? true : Object.keys(current);
}
