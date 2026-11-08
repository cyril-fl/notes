<script setup lang="ts">
import type { User } from '~/types/items';

const route = useRoute();
const api = useApi();
const { setLocale } = useI18n();
const { createFolder } = useItems();
const searchOpen = useState('search-open', () => false);
const searchSnippets = useState('search-snippets', () => false);
const collapsed = ref(false);

// The focus mode leaves only the note: it ends with the page, or with the Pro mode.
const { has } = usePro();
const focus = useState('focus', () => false);
const focused = computed(() => focus.value && has('focusMode'));
watch(
  () => route.path,
  () => (focus.value = false),
);
// The note's editor fills it (a Teleport to #context-panel).
const contextPanel = useContextPanel();
const notePanel = computed(() => contextPanel.value && route.path.startsWith('/notes/'));

// The account's language wins over the cookie's, which only remembers this browser's last one.
const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));
watch(
  () => me.value?.locale,
  (locale) => locale && setLocale(locale as 'fr' | 'en'),
  { immediate: true }, // me may be loaded already, by the settings page's middleware
);

// Theme, mode and view follow the account; a change of mode or view from anywhere (sidebar, view toggle, settings) is saved to it.
const appConfig = useAppConfig();
const colorMode = useColorMode();
const view = useView();
watch(
  me,
  (user) => {
    if (!user) return;
    const theme = themes.find((theme) => theme.id === user.theme) ?? themes[0]!;
    appConfig.ui.colors.primary = theme.primary;
    appConfig.ui.colors.neutral = theme.neutral;
    colorMode.preference = user.color_mode;
    view.value = user.view as 'list' | 'grid';
  },
  { immediate: true },
);
watch([() => colorMode.preference, view], async ([mode, next]) => {
  if (!me.value || (mode === me.value.color_mode && next === me.value.view)) return;
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { color_mode: mode, view: next } });
});

// Single keys, like Notion or Linear: Nuxt UI skips them while typing in a field or the editor.
defineShortcuts({
  '/': () => (searchOpen.value = true),
  n: () => navigateTo(newNotePath(currentFolder(route))),
  shift_n: () => createFolder(currentFolder(route)),
  t: () => (searchSnippets.value = searchOpen.value = true),
  'meta_\\': () => (collapsed.value = !collapsed.value),
  'meta_.': { usingInput: true, handler: () => has('focusMode') && (focus.value = !focus.value) },
});
// The template: Nuxt UI's dashboard, a sidebar that resizes (its width kept in localStorage), collapses, and becomes a drawer on small screens.
</script>

<template>
  <UDashboardGroup unit="rem" storage="local" storage-key="notes-layout" class="gap-2 bg-muted p-2 text-default">
    <UDashboardSidebar
      v-if="!focused"
      v-model:collapsed="collapsed"
      collapsible
      resizable
      :min-size="12"
      :default-size="16"
      :max-size="28"
      :collapsed-size="3.5"
      :ui="{
        root: 'h-full min-h-0 min-w-0 overflow-hidden rounded-xl border-0 bg-default',
        body: 'gap-0 p-0',
      }"
    >
      <template #default="{ collapsed: rail }">
        <AppSidebar :collapsed="rail" />
      </template>
    </UDashboardSidebar>

    <UDashboardPanel
      :ui="{
        root: 'h-full min-h-0 flex-1 overflow-hidden rounded-xl border-0 bg-default',
        body: 'gap-0 p-0 sm:gap-0 sm:p-0',
      }"
    >
      <template #body>
        <!-- Only on a small screen, where the sidebar is a drawer. -->
        <div v-if="!focused" class="flex px-2 pt-2 lg:hidden">
          <UDashboardSidebarToggle color="neutral" variant="ghost" />
        </div>
        <main class="flex min-h-0 flex-1 flex-col">
          <slot />
        </main>
      </template>
    </UDashboardPanel>

    <!-- Not collapsible: the sidebar's collapse button would fold it too. -->
    <UDashboardSidebar
      v-if="notePanel"
      id="context"
      side="right"
      resizable
      :min-size="12"
      :default-size="16"
      :max-size="28"
      :ui="{
        root: 'h-full min-h-0 min-w-0 overflow-hidden rounded-xl border-0 bg-default',
        body: 'gap-0 p-0',
      }"
    >
      <aside id="context-panel" class="flex flex-1 flex-col gap-6 p-4" :aria-label="$t('features.contextPanel')" />
    </UDashboardSidebar>

    <AppSearch />
  </UDashboardGroup>
</template>
