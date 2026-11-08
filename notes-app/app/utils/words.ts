const words = new Intl.Segmenter(undefined, { granularity: 'word' });

/** The words of a text, as the browser cuts them: punctuation and emoji are not words, each CJK word is one. */
export function countWords(text: string) {
  let count = 0;
  for (const segment of words.segment(text)) if (segment.isWordLike) count++;
  return count;
}
