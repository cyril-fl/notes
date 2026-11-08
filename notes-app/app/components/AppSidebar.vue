<script setup lang="ts">
import type { Recent, TagCount, User } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

defineProps<{ collapsed?: boolean }>();

const api = useApi();
const route = useRoute();
const { createFolder, tagMenu, pinItem } = useItems();
const searchOpen = useState('search-open', () => false);
const colorMode = useColorMode();
const { has } = usePro();
// Pro: on the settings, the sidebar becomes their nav, like Kaneo's.
const settingsNav = computed(
  () => has('settingsNav') && (route.path.startsWith('/settings') || route.path === '/account'),
);
const settingsLinks = computed(() => [
  { label: t('sidebar.account'), icon: icons.account, to: '/account' },
  ...settingsSections.map((section) => ({
    label: t(`settings.sections.${section.id}`),
    icon: icons[section.icon],
    to: `/settings/${section.id}`,
  })),
]);

const { data: pinned } = usePinned();
const favorites = computed(() => pinned.value.filter((item) => item.kind === 'folder'));
const pinnedNotes = computed(() => pinned.value.filter((item) => item.kind === 'note'));
// In system mode the OS decides: the sidebar shows no toggle that would override it.
const darkMode = computed({
  get: () => colorMode.value === 'dark',
  set: (dark) => (colorMode.preference = dark ? 'dark' : 'light'),
});
const { data: recent } = useLazyAsyncData('recent', () => api<Recent[]>('/api/recent', { quiet: true }), {
  default: () => [] as Recent[],
});
const { data: tags } = useLazyAsyncData('tags', () => api<TagCount[]>('/api/graph'), {
  default: () => [] as TagCount[],
});

// The account, loaded by the layout: its avatar opens the account menu.
const { data: me } = useNuxtData<User>('me');
const baseURL = useRuntimeConfig().public.apiBase;
const avatarSrc = computed(() => (me.value?.avatar ? imageSrc(`/api/files/${me.value.avatar}`, baseURL) : undefined));
const logout = useLogout();
const accountMenu = computed(() => [
  [{ label: me.value?.name, type: 'label' as const }],
  // The account is reached from the settings, which link to it: not repeated here.
  [
    { label: t('common.settings'), icon: icons.settings, to: '/settings' },
    { label: t('settings.logout'), icon: icons.logout, onSelect: logout },
  ],
]);

const rootLink = useTemplateRef('rootLink');
const isOverRoot = useDropFolder(rootLink, () => null);
</script>

<template>
  <!-- Reduced to a rail of icons: the sidebar never leaves the screen. -->
  <nav v-if="collapsed" class="flex h-full flex-col items-center gap-1 py-3" :aria-label="$t('common.navigation')">
    <UDropdownMenu :items="accountMenu" :content="{ side: 'right', align: 'start' }">
      <UButton color="neutral" variant="ghost" :aria-label="$t('sidebar.accountMenu')">
        <UAvatar :src="avatarSrc" :alt="me?.name" size="xs" />
      </UButton>
    </UDropdownMenu>
    <UTooltip :text="$t('common.search')" :kbds="['meta', 'K']" :content="{ side: 'right' }">
      <UButton
        :icon="icons.search"
        color="neutral"
        variant="ghost"
        :aria-label="$t('common.search')"
        @click="searchOpen = true"
      />
    </UTooltip>
    <template v-if="settingsNav">
      <UTooltip :text="$t('common.backToNotes')" :content="{ side: 'right' }">
        <UButton :icon="icons.back" to="/" color="neutral" variant="ghost" :aria-label="$t('common.backToNotes')" />
      </UTooltip>
      <UTooltip v-for="link in settingsLinks" :key="link.to" :text="link.label" :content="{ side: 'right' }">
        <UButton
          :icon="link.icon"
          :to="link.to"
          color="neutral"
          variant="ghost"
          active-variant="soft"
          :aria-label="link.label"
        />
      </UTooltip>
    </template>
    <template v-else>
      <UTooltip :text="$t('common.notes')" :content="{ side: 'right' }">
        <UButton
          :icon="icons.notes"
          to="/"
          color="neutral"
          variant="ghost"
          active-variant="soft"
          :aria-label="$t('common.notes')"
          :active="route.path === '/'"
        />
      </UTooltip>
      <UTooltip v-if="has('graph')" :text="$t('common.graph')" :content="{ side: 'right' }">
        <UButton
          :icon="icons.graph"
          to="/graph"
          color="neutral"
          variant="ghost"
          active-variant="soft"
          :aria-label="$t('common.graph')"
        />
      </UTooltip>
    </template>

    <div class="mt-auto flex flex-col items-center gap-1 border-t border-default pt-2">
      <UTooltip :text="$t('common.trash')" :content="{ side: 'right' }">
        <UButton
          :icon="icons.trash"
          to="/trash"
          color="neutral"
          variant="ghost"
          active-variant="soft"
          :aria-label="$t('common.trash')"
        />
      </UTooltip>
      <UTooltip
        v-if="colorMode.preference !== 'system'"
        :text="darkMode ? $t('sidebar.lightTheme') : $t('sidebar.darkTheme')"
        :content="{ side: 'right' }"
      >
        <UButton
          :icon="darkMode ? icons.lightMode : icons.darkMode"
          color="neutral"
          variant="ghost"
          :aria-label="darkMode ? $t('sidebar.lightTheme') : $t('sidebar.darkTheme')"
          @click="darkMode = !darkMode"
        />
      </UTooltip>
      <UTooltip :text="$t('sidebar.expand')" :kbds="['meta', '\\']" :content="{ side: 'right' }">
        <UDashboardSidebarCollapse color="neutral" variant="ghost" />
      </UTooltip>
    </div>
  </nav>

  <nav v-else class="flex h-full flex-col gap-4 p-3" :aria-label="$t('common.navigation')">
    <div class="flex items-center gap-1">
      <UDropdownMenu :items="accountMenu" :content="{ side: 'bottom', align: 'start' }">
        <UButton color="neutral" variant="ghost" size="sm" :aria-label="$t('sidebar.accountMenu')">
          <UAvatar :src="avatarSrc" :alt="me?.name" size="2xs" />
        </UButton>
      </UDropdownMenu>
      <UButton
        :icon="icons.search"
        :label="$t('common.search')"
        color="neutral"
        variant="soft"
        size="sm"
        block
        class="min-w-0 flex-1 justify-start text-muted"
        @click="searchOpen = true"
      >
        <template #trailing> <UKbd value="meta" size="sm" class="ms-auto" /><UKbd value="K" size="sm" /> </template>
      </UButton>
    </div>

    <div v-if="settingsNav" class="min-h-0 flex-1 overflow-y-auto">
      <NuxtLink to="/" class="row font-medium">
        <UIcon :name="icons.back" class="size-4 shrink-0" />
        {{ $t('common.backToNotes') }}
      </NuxtLink>
      <h2 class="px-2 pt-4 pb-1 text-xs font-medium text-muted">{{ $t('common.settings') }}</h2>
      <NuxtLink
        v-for="link in settingsLinks"
        :key="link.to"
        :to="link.to"
        class="row"
        :class="{ 'row-active': route.path === link.to }"
      >
        <UIcon :name="link.icon" class="size-4 shrink-0" />
        {{ link.label }}
      </NuxtLink>
    </div>
    <template v-else>
      <!-- Above the scrolling area, so a long folder tree never hides it. -->
      <NuxtLink
        v-if="has('graph')"
        to="/graph"
        class="row -mt-2 font-medium"
        :class="{ 'row-active': route.path === '/graph' }"
      >
        <UIcon :name="icons.graph" class="size-4 shrink-0" />
        {{ $t('common.graph') }}
      </NuxtLink>

      <div class="min-h-0 flex-1 space-y-5 overflow-y-auto">
        <section v-if="favorites.length">
          <h2 class="px-2 pb-1 text-xs font-medium text-muted">{{ $t('sidebar.favorites') }}</h2>
          <div
            v-for="folder in favorites"
            :key="folder.id"
            class="row"
            :class="{ 'row-active': route.params.id === folder.id }"
          >
            <UTooltip :text="$t('sidebar.removeFavorite')">
              <button
                type="button"
                class="shrink-0 text-muted hover:text-default"
                :aria-label="$t('sidebar.removeFavorite')"
                @click="pinItem(folder, false)"
              >
                <UIcon :name="icons.favorite" class="block size-4" />
              </button>
            </UTooltip>
            <NuxtLink :to="`/folders/${folder.id}`" class="flex-1 self-stretch truncate leading-7">{{
              folder.title
            }}</NuxtLink>
          </div>
        </section>

        <section v-if="pinnedNotes.length">
          <h2 class="px-2 pb-1 text-xs font-medium text-muted">{{ $t('sidebar.pinned') }}</h2>
          <div
            v-for="note in pinnedNotes"
            :key="note.id"
            class="row"
            :class="{ 'row-active': route.params.id === note.id }"
          >
            <UTooltip :text="$t('sidebar.unpin')">
              <button
                type="button"
                class="shrink-0 text-muted hover:text-default"
                :aria-label="$t('sidebar.unpin')"
                @click="pinItem(note, false)"
              >
                <UIcon :name="icons.pin" class="block size-4" />
              </button>
            </UTooltip>
            <NuxtLink :to="`/notes/${note.id}`" class="flex-1 self-stretch truncate leading-7">{{
              displayTitle(note.title, $t('untitled'))
            }}</NuxtLink>
          </div>
        </section>

        <section v-if="recent.length">
          <h2 class="px-2 pb-1 text-xs font-medium text-muted">{{ $t('sidebar.recent') }}</h2>
          <NuxtLink
            v-for="note in recent.slice(0, 5)"
            :key="note.id"
            :to="`/notes/${note.id}`"
            class="row"
            :class="{ 'row-active': route.params.id === note.id }"
          >
            <UIcon :name="icons.recent" class="size-4 shrink-0 text-muted" />
            <span class="truncate">{{ displayTitle(note.title, $t('untitled')) }}</span>
          </NuxtLink>
        </section>

        <section>
          <div class="flex items-center">
            <NuxtLink
              ref="rootLink"
              to="/"
              class="row flex-1 font-medium"
              :class="{ 'row-active': route.path === '/', 'row-over': isOverRoot }"
            >
              <UIcon :name="icons.notes" class="size-4 shrink-0" />
              {{ $t('common.notes') }}
            </NuxtLink>
            <UTooltip :text="$t('common.newFolder')">
              <UButton
                :icon="icons.folderAdd"
                color="neutral"
                variant="ghost"
                size="xs"
                :aria-label="$t('common.newFolder')"
                @click="createFolder()"
              />
            </UTooltip>
          </div>
          <FolderTree />
        </section>
      </div>
    </template>

    <section v-if="!settingsNav && tags.length" class="border-t border-default pt-3">
      <h2 class="px-2 pb-2 text-xs font-medium text-muted">{{ $t('common.tags') }}</h2>
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
        {{ $t('common.trash') }}
      </NuxtLink>
      <UTooltip
        v-if="colorMode.preference !== 'system'"
        :text="darkMode ? $t('sidebar.lightTheme') : $t('sidebar.darkTheme')"
      >
        <UButton
          :icon="darkMode ? icons.lightMode : icons.darkMode"
          color="neutral"
          variant="ghost"
          size="sm"
          :aria-label="darkMode ? $t('sidebar.lightTheme') : $t('sidebar.darkTheme')"
          @click="darkMode = !darkMode"
        />
      </UTooltip>
      <UTooltip :text="$t('sidebar.collapse')" :kbds="['meta', '\\']">
        <UDashboardSidebarCollapse color="neutral" variant="ghost" size="sm" class="hidden lg:inline-flex" />
      </UTooltip>
    </div>
  </nav>
</template>
