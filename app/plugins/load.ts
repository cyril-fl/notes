export default defineNuxtPlugin({
  name: 'load',
  setup() {
    const { loadData } = useDataStore();
    const { data } = useAsyncData<ServerResponse<DataSchema[]>>('data', () =>
      $fetch<ServerResponse<DataSchema[]>>('/api/data/read')
    );

    watch(
      data,
      (newValue) => {
        if (!newValue) return;
        console.log('🚀 [PLUGIN] - Loading data...', newValue);
        loadData(newValue.data);
      },
      { immediate: true }
    );
  },
});
