/** The items selected with ⌘ or ⇧ in the folder shown: emptied when the page changes. */
export const useSelection = createGlobalState(() => {
  const ids = ref<string[]>([]);
  const route = useRoute();
  watch(() => route.fullPath, () => (ids.value = []));
  return ids;
});
