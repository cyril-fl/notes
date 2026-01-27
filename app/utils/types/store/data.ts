interface ExtractProp {
  note: NotePublic;
  tags: MappedTags;
}
export const extractTags = ({ note, tags }: ExtractProp) => {
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
