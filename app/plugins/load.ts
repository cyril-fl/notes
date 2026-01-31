export default defineNuxtPlugin(async () => {
  const { onRead } = useDataStore();
  const { data } = useAsyncData<ServerResponse<DataSchema[]>>('data', () =>
    $fetch<ServerResponse<DataSchema[]>>('/api/data/read')
  );

  watch(
    data,
    (newValue) => {
      if (!newValue) return;
      console.log('🚀 [PLUGIN] - Loading data...', newValue);
      onRead(newValue.data);
    },
    { immediate: true }
  );
});
