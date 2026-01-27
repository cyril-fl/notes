export const useMentions = (data?: string[]) => {
  const mentions = ref<string[]>(data ?? []);

  const handleUpdate = (newValue: string[]) => {
    mentions.value = newValue;
    console.log('[MENTIONS] update:', newValue);
  };

  return {
    mentions,
    onUpdate: handleUpdate,
  };
};
