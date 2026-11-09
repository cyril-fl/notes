import { diff3Merge, diffComm } from 'node-diff3';

// Lines both sides agree on, or a conflict: lines changed on both sides in the same place.
export type MergePart = { lines: string[] } | { mine: string[]; theirs: string[] };

/** A three-way merge by lines (#78): base is the text both sides started from. */
export function merge(base: string, mine: string, theirs: string): MergePart[] {
  return diff3Merge(mine.split('\n'), base.split('\n'), theirs.split('\n')).map((region) =>
    region.conflict ? { mine: region.conflict.a, theirs: region.conflict.b } : { lines: region.ok ?? [] },
  );
}

export type Choice = 'mine' | 'theirs' | 'both';

/** The text of the parts, each conflict as chosen. */
export function mergedText(parts: MergePart[], choices: Choice[] = []) {
  let conflict = 0;
  return parts
    .flatMap((part) => {
      if ('lines' in part) return part.lines;
      const choice = choices[conflict++];
      // Both: a blank line between, or two paragraphs would read as one.
      return choice === 'mine' ? part.mine : choice === 'theirs' ? part.theirs : [...part.mine, '', ...part.theirs];
    })
    .join('\n');
}

export type Change = { kind: 'same' | 'removed' | 'added'; lines: string[] };

/** The lines kept, removed and added from before to after, for the comparison of a version (#78). */
export function lineChanges(before: string, after: string): Change[] {
  return diffComm(before.split('\n'), after.split('\n')).flatMap((block): Change[] =>
    block.common
      ? [{ kind: 'same', lines: block.common }]
      : [
          { kind: 'removed' as const, lines: block.buffer1 },
          { kind: 'added' as const, lines: block.buffer2 },
        ].filter((change) => change.lines.length),
  );
}
