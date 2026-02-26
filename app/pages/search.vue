<script setup lang="ts">
const route = useRoute();
const { t } = useI18n();

usePageSection({ title: t('pages.search.title') });

const results = ref<(Note | Folder)[]>([]);
const loading = ref(false);
const hasSearched = ref(false);

function toInstance(item: DataSchema): Note | Folder {
  if (item.type === ItemType.NOTE) {
    return new Note(noteParams.parse(item));
  }
  return new Folder(folderParams.parse(item));
}

async function performSearch() {
  const q = route.query.q as string;
  if (!q) return;

  loading.value = true;
  hasSearched.value = true;

  try {
    const response = await $fetch<ServerResponse<DataSchema[]>>(
      '/api/data/search',
      {
        query: {
          q,
          type: route.query.type as string,
          tags: route.query.tags as string,
        },
      }
    );

    results.value = response.data.map(toInstance);
  } catch {
    results.value = [];
  } finally {
    loading.value = false;
  }
}

watch(() => route.query, performSearch, { immediate: true });
</script>

<template>
  <div v-if="loading" class="text-muted text-sm">
    {{ t('pages.search.loading') }}
  </div>

  <div
    v-else-if="hasSearched && results.length === 0"
    class="text-muted text-sm"
  >
    {{ t('pages.search.empty') }}
  </div>

  <ul v-else class="flex flex-wrap gap-4">
    <li v-for="item in results" :key="item.id">
      <UINotesCard v-if="item.type === ItemType.NOTE" :item="item as Note" />
      <UIFolderCard v-else :item="item as Folder" />
    </li>
  </ul>
</template>
