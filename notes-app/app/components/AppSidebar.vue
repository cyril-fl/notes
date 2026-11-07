<script setup lang="ts">
import type { Recent, TagCount } from '~/types/items';

const api = useApi();
const route = useRoute();
const { createFolder, tagMenu } = useItems();
const searchOpen = useState('search-open', () => false);
const colorMode = useColorMode();

const { data: pinned } = usePinned();
const favorites = computed(() => pinned.value.filter((item) => item.kind === 'folder'));
const pinnedNotes = computed(() => pinned.value.filter((item) => item.kind === 'note'));
const { data: recent } = useLazyAsyncData('recent', () => api<Recent[]>('/api/recent', { quiet: true }), { default: () => [] as Recent[] });
const { data: tags } = useLazyAsyncData('tags', () => api<TagCount[]>('/api/graph'), { default: () => [] as TagCount[] });

const rootLink = useTemplateRef('rootLink');
const isOverRoot = useDropFolder(rootLink, () => null);
</script>

<template>
  <nav class="flex h-full flex-col gap-4 p-3" aria-label="Navigation">
    <UButton
      icon="i-lucide-search"
      label="Rechercher"
      color="neutral"
      variant="soft"
      size="sm"
      block
      class="justify-start text-muted"
      @click="searchOpen = true"
    >
      <template #trailing>
        <UKbd value="meta" size="sm" class="ms-auto" /><UKbd value="K" size="sm" />
      </template>
    </UButton>

    <div class="min-h-0 flex-1 space-y-5 overflow-y-auto">
      <section v-if="favorites.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Favoris</h2>
        <NuxtLink v-for="folder in favorites" :key="folder.id" :to="`/folders/${folder.id}`" class="row" :class="{ 'row-active': route.params.id === folder.id }">
          <UIcon name="i-lucide-star" class="size-4 shrink-0 text-muted" />
          <span class="truncate">{{ folder.title }}</span>
        </NuxtLink>
      </section>

      <section v-if="pinnedNotes.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Épinglées</h2>
        <NuxtLink v-for="note in pinnedNotes" :key="note.id" :to="`/notes/${note.id}`" class="row" :class="{ 'row-active': route.params.id === note.id }">
          <UIcon name="i-lucide-pin" class="size-4 shrink-0 text-muted" />
          <span class="truncate">{{ displayTitle(note.title) }}</span>
        </NuxtLink>
      </section>

      <section v-if="recent.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Récentes</h2>
        <NuxtLink v-for="note in recent.slice(0, 5)" :key="note.id" :to="`/notes/${note.id}`" class="row" :class="{ 'row-active': route.params.id === note.id }">
          <UIcon name="i-lucide-clock" class="size-4 shrink-0 text-muted" />
          <span class="truncate">{{ displayTitle(note.title) }}</span>
        </NuxtLink>
      </section>

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
    </div>

    <section v-if="tags.length" class="border-t border-default pt-3">
      <h2 class="px-2 pb-2 text-xs font-medium text-muted">Tags</h2>
      <!-- Capped so a long tag list never pushes the folders out of the sidebar. -->
      <div class="flex max-h-40 flex-wrap gap-1.5 overflow-y-auto px-2">
        <UContextMenu v-for="tag in tags" :key="tag.value" :items="tagMenu(tag.value)">
          <UButton
            :to="tagPath(tag.value)"
            :label="`#${tag.value}`"
            size="xs"
            color="neutral"
            variant="soft"
            active-color="primary"
            class="max-w-full font-normal"
            :ui="{ label: 'truncate' }"
          />
        </UContextMenu>
      </div>
    </section>

    <div class="flex items-center justify-between border-t border-default pt-2">
      <NuxtLink to="/trash" class="row flex-1" :class="{ 'row-active': route.path === '/trash' }">
        <UIcon name="i-lucide-trash-2" class="size-4 shrink-0" />
        Corbeille
      </NuxtLink>
      <UTooltip text="Réglages">
        <UButton icon="i-lucide-settings" to="/settings" color="neutral" variant="ghost" size="sm" aria-label="Réglages" />
      </UTooltip>
      <UTooltip :text="colorMode.value === 'dark' ? 'Thème clair' : 'Thème sombre'">
        <UButton
          :icon="colorMode.value === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="colorMode.value === 'dark' ? 'Thème clair' : 'Thème sombre'"
          @click="colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'"
        />
      </UTooltip>
      <UTooltip text="Nouveau dossier">
        <UButton icon="i-lucide-folder-plus" color="neutral" variant="ghost" size="sm" aria-label="Nouveau dossier" @click="createFolder()" />
      </UTooltip>
      <UTooltip text="Replier la barre latérale" :kbds="['meta', '\\']">
        <UDashboardSidebarCollapse color="neutral" variant="ghost" size="sm" class="hidden lg:inline-flex" />
      </UTooltip>
    </div>
  </nav>
</template>
