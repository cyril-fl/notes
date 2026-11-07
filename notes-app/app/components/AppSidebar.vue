<script setup lang="ts">
import type { Recent, TagCount } from '~/types/items';

const { icons } = useIcons();

defineProps<{ collapsed?: boolean }>();

const api = useApi();
const route = useRoute();
const { createFolder, tagMenu, pinItem } = useItems();
const searchOpen = useState('search-open', () => false);
const colorMode = useColorMode();

const { data: pinned } = usePinned();
const favorites = computed(() => pinned.value.filter((item) => item.kind === 'folder'));
const pinnedNotes = computed(() => pinned.value.filter((item) => item.kind === 'note'));
const darkMode = computed({
  get: () => colorMode.value === 'dark',
  set: (dark) => (colorMode.preference = dark ? 'dark' : 'light'),
});
const { data: recent } = useLazyAsyncData('recent', () => api<Recent[]>('/api/recent', { quiet: true }), { default: () => [] as Recent[] });
const { data: tags } = useLazyAsyncData('tags', () => api<TagCount[]>('/api/graph'), { default: () => [] as TagCount[] });

const rootLink = useTemplateRef('rootLink');
const isOverRoot = useDropFolder(rootLink, () => null);
</script>

<template>
  <!-- Reduced to a rail of icons: the sidebar never leaves the screen. -->
  <nav v-if="collapsed" class="flex h-full flex-col items-center gap-1 py-3" aria-label="Navigation">
    <UTooltip text="Rechercher" :kbds="['meta', 'K']" :content="{ side: 'right' }">
      <UButton :icon="icons.search" color="neutral" variant="ghost" aria-label="Rechercher" @click="searchOpen = true" />
    </UTooltip>
    <UTooltip text="Notes" :content="{ side: 'right' }">
      <UButton :icon="icons.notes" to="/" color="neutral" variant="ghost" active-variant="soft" aria-label="Notes" :active="route.path === '/'" />
    </UTooltip>
    <UTooltip text="Graphe" :content="{ side: 'right' }">
      <UButton :icon="icons.graph" to="/graph" color="neutral" variant="ghost" active-variant="soft" aria-label="Graphe" />
    </UTooltip>
    <UTooltip text="Corbeille" :content="{ side: 'right' }">
      <UButton :icon="icons.trash" to="/trash" color="neutral" variant="ghost" active-variant="soft" aria-label="Corbeille" />
    </UTooltip>

    <div class="mt-auto flex flex-col items-center gap-1 border-t border-default pt-2">
      <UTooltip text="Réglages" :content="{ side: 'right' }">
        <UButton :icon="icons.settings" to="/settings" color="neutral" variant="ghost" active-variant="soft" aria-label="Réglages" />
      </UTooltip>
      <UTooltip :text="darkMode ? 'Thème clair' : 'Thème sombre'" :content="{ side: 'right' }">
        <UButton :icon="darkMode ? icons.lightMode : icons.darkMode" color="neutral" variant="ghost" :aria-label="darkMode ? 'Thème clair' : 'Thème sombre'" @click="darkMode = !darkMode" />
      </UTooltip>
      <UTooltip text="Déployer la barre latérale" :kbds="['meta', '\\']" :content="{ side: 'right' }">
        <UDashboardSidebarCollapse color="neutral" variant="ghost" />
      </UTooltip>
    </div>
  </nav>

  <nav v-else class="flex h-full flex-col gap-4 p-3" aria-label="Navigation">
    <UButton
      :icon="icons.search"
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
        <div v-for="folder in favorites" :key="folder.id" class="row" :class="{ 'row-active': route.params.id === folder.id }">
          <UTooltip text="Retirer des favoris">
            <button type="button" class="shrink-0 text-muted hover:text-default" aria-label="Retirer des favoris" @click="pinItem(folder, false)">
              <UIcon :name="icons.favorite" class="block size-4" />
            </button>
          </UTooltip>
          <NuxtLink :to="`/folders/${folder.id}`" class="flex-1 self-stretch truncate leading-7">{{ folder.title }}</NuxtLink>
        </div>
      </section>

      <section v-if="pinnedNotes.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Épinglées</h2>
        <div v-for="note in pinnedNotes" :key="note.id" class="row" :class="{ 'row-active': route.params.id === note.id }">
          <UTooltip text="Désépingler">
            <button type="button" class="shrink-0 text-muted hover:text-default" aria-label="Désépingler" @click="pinItem(note, false)">
              <UIcon :name="icons.pin" class="block size-4" />
            </button>
          </UTooltip>
          <NuxtLink :to="`/notes/${note.id}`" class="flex-1 self-stretch truncate leading-7">{{ displayTitle(note.title) }}</NuxtLink>
        </div>
      </section>

      <section v-if="recent.length">
        <h2 class="px-2 pb-1 text-xs font-medium text-muted">Récentes</h2>
        <NuxtLink v-for="note in recent.slice(0, 5)" :key="note.id" :to="`/notes/${note.id}`" class="row" :class="{ 'row-active': route.params.id === note.id }">
          <UIcon :name="icons.recent" class="size-4 shrink-0 text-muted" />
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
          <UIcon :name="icons.notes" class="size-4 shrink-0" />
          Notes
        </NuxtLink>
        <FolderTree />
      </section>

      <NuxtLink to="/graph" class="row font-medium" :class="{ 'row-active': route.path === '/graph' }">
        <UIcon :name="icons.graph" class="size-4 shrink-0" />
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
        <UIcon :name="icons.trash" class="size-4 shrink-0" />
        Corbeille
      </NuxtLink>
      <UTooltip text="Réglages">
        <UButton :icon="icons.settings" to="/settings" color="neutral" variant="ghost" size="sm" aria-label="Réglages" />
      </UTooltip>
      <UTooltip :text="darkMode ? 'Thème clair' : 'Thème sombre'">
        <UButton :icon="darkMode ? icons.lightMode : icons.darkMode" color="neutral" variant="ghost" size="sm" :aria-label="darkMode ? 'Thème clair' : 'Thème sombre'" @click="darkMode = !darkMode" />
      </UTooltip>
      <UTooltip text="Nouveau dossier">
        <UButton :icon="icons.folderAdd" color="neutral" variant="ghost" size="sm" aria-label="Nouveau dossier" @click="createFolder()" />
      </UTooltip>
      <UTooltip text="Replier la barre latérale" :kbds="['meta', '\\']">
        <UDashboardSidebarCollapse color="neutral" variant="ghost" size="sm" class="hidden lg:inline-flex" />
      </UTooltip>
    </div>
  </nav>
</template>
