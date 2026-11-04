<script setup lang="ts">
import type { TagCount } from '~/types/items';

const api = useApi();
const route = useRoute();
const { createFolder } = useItems();

const search = ref(typeof route.query.q === 'string' ? route.query.q : '');
watchDebounced(search, (q) => {
  if (q.trim()) navigateTo({ path: '/search', query: { q } });
}, { debounce: 250 });

const { data: tags } = useLazyAsyncData('tags', () => api<TagCount[]>('/api/graph'), { default: () => [] as TagCount[] });

const rootLink = useTemplateRef('rootLink');
const isOverRoot = useDropFolder(rootLink, () => null);
</script>

<template>
  <nav class="flex h-full flex-col gap-4 p-3" aria-label="Navigation">
    <UInput
      v-model="search"
      icon="i-lucide-search"
      placeholder="Rechercher"
      aria-label="Rechercher"
      variant="soft"
      size="sm"
    />

    <div class="min-h-0 flex-1 space-y-5 overflow-y-auto">
      <section>
        <NuxtLink
          ref="rootLink"
          to="/"
          class="row font-medium"
          :class="{ 'row-active': route.path === '/', 'row-over': isOverRoot }"
        >
          <UIcon name="i-lucide-notebook-text" class="size-4 shrink-0" />
          Notes
        </NuxtLink>
        <FolderTree />
      </section>

      <NuxtLink to="/graph" class="row font-medium" :class="{ 'row-active': route.path === '/graph' }">
        <UIcon name="i-lucide-waypoints" class="size-4 shrink-0" />
        Graphe
      </NuxtLink>

      <section v-if="tags.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Tags</h2>
        <NuxtLink
          v-for="tag in tags"
          :key="tag.value"
          :to="`/tags/${encodeURIComponent(tag.value)}`"
          class="row"
          :class="{ 'row-active': route.params.value === tag.value }"
        >
          <span class="text-muted">#</span>
          <span class="truncate">{{ tag.value }}</span>
          <span class="ml-auto text-xs text-muted">{{ tag.notes }}</span>
        </NuxtLink>
      </section>
    </div>

    <div class="flex items-center justify-between border-t border-default pt-2">
      <NuxtLink to="/trash" class="row flex-1" :class="{ 'row-active': route.path === '/trash' }">
        <UIcon name="i-lucide-trash-2" class="size-4 shrink-0" />
        Corbeille
      </NuxtLink>
      <UTooltip text="Nouveau dossier">
        <UButton icon="i-lucide-folder-plus" color="neutral" variant="ghost" size="sm" aria-label="Nouveau dossier" @click="createFolder()" />
      </UTooltip>
    </div>
  </nav>
</template>
