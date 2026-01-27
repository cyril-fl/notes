export const useHashtags = (data?: string[]) => {
  const hashtags = ref<string[]>(data ?? []);

  const handleUpdate = (newValue: string[]) => {
    hashtags.value = newValue;
    console.log('[HASHTAGS] update:', newValue);
  };

  return {
    hashtags,
    onUpdate: handleUpdate,
  };
};
