<script setup lang="ts">
import type { SavedSearchResult, User } from '~/types/items';

definePageMeta({ feature: 'savedSearches' });

const { icons } = useIcons();

const route = useRoute();
const id = route.params.id as string;
const api = useApi();
const view = useView();
const { editSavedSearch, removeSavedSearch } = useSavedSearches();

const { data: results, error } = useLazyAsyncData(`search:${id}`, () =>
  api<SavedSearchResult[]>(`/api/searches/${id}`, { quiet: true }),
);
const search = computed(() => useNuxtData<User>('me').data.value?.saved_searches.find((s) => s.id === id));
useHead({ title: () => search.value?.name ?? '' });
</script>

<template>
  <div class="p-8">
    <PageHeader :title="search?.name ?? ''" :subtitle="results ? $t('tag.count', results.length) : undefined">
      <ViewToggle />
      <UTooltip v-if="search" :text="$t('common.edit')">
        <UButton
          :icon="icons.rename"
          color="neutral"
          variant="ghost"
          :aria-label="$t('common.edit')"
          @click="editSavedSearch(search)"
        />
      </UTooltip>
      <UTooltip v-if="search" :text="$t('common.delete')">
        <UButton
          :icon="icons.delete"
          color="neutral"
          variant="ghost"
          :aria-label="$t('common.delete')"
          @click="removeSavedSearch(search)"
        />
      </UTooltip>
    </PageHeader>
    <NoteList v-if="results?.length" :notes="results" :grid="view === 'grid'" />
    <UEmpty
      v-else-if="error"
      :icon="icons.search"
      :title="$t('search.saved.notFound')"
      :description="$t('search.saved.notFoundDescription')"
      :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
    />
    <UEmpty
      v-else-if="results"
      :icon="icons.searchEmpty"
      :title="$t('search.saved.empty')"
      :description="$t('search.saved.emptyDescription')"
    />
  </div>
</template>
