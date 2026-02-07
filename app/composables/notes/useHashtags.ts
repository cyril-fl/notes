export const useHashtags = () => {
  const store = useDataStore();
  const { data } = storeToRefs(store);

  const hashtags = computed(() => {
    const raw = data.value.notes.map((note) => note.hashtags).flat();
    const set = new Set(raw);

    return Array.from(set);
  });

  return {
    hashtags,
    hasTags: hashtags.value.length > 0,
  };
};
