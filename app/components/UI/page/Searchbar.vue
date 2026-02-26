<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui';
import {
  SEARCH_TYPE_ALL,
  searchTypes,
  type SearchType,
} from '~~/shared/types/api';

const { icons } = useIcons();
const route = useRoute();
const router = useRouter();

const selectedType = ref<SearchType>(SEARCH_TYPE_ALL);

const queryText = ref('');

// Sync from route query when on search page (browser refresh)
if (route.path === '/search') {
  queryText.value = (route.query.q as string) || '';
  const routeType = route.query.type as string;
  if (
    routeType &&
    (searchTypes as readonly string[]).includes(routeType)
  ) {
    selectedType.value = routeType as SearchType;
  }
}

const performSearch = useDebounceFn(() => {
  if (!queryText.value.trim()) {
    if (route.path === '/search') router.back();
    return;
  }

  navigateTo({
    path: '/search',
    query: {
      q: queryText.value.trim(),
      type: selectedType.value,
    },
  });
}, 300);

watch(selectedType, () => {
  if (queryText.value.trim()) performSearch();
});

const items = computed<DropdownMenuItem[]>(() => [
  {
    label: `Types: ${selectedType.value}`,
    icon: icons.note,
    onSelect(e: Event) {
      e.preventDefault();
      const currentIndex = searchTypes.indexOf(selectedType.value);
      const nextIndex = (currentIndex + 1) % searchTypes.length;
      selectedType.value = searchTypes[nextIndex]!;
    },
  },
]);
</script>

<template>
  <UFieldGroup>
    <UInput
      v-model="queryText"
      :icon="icons.search"
      :placeholder="$t('pages.search.placeholder')"
      @input="performSearch"
    />
    <UDropdownMenu :items="items">
      <UButton :icon="icons.filter" variant="outline" color="neutral" />
    </UDropdownMenu>
  </UFieldGroup>
</template>
