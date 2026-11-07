<script setup lang="ts">
import type { User } from '~/types/items';

const route = useRoute();
const api = useApi();
const { setLocale } = useI18n();
const { createFolder } = useItems();
const searchOpen = useState('search-open', () => false);
const collapsed = ref(false);

// The account's language wins over the cookie's, which only remembers this browser's last one.
const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));
watch(() => me.value?.locale, (locale) => locale && setLocale(locale as 'fr' | 'en'));

// Theme and view follow the account; a change from anywhere (sidebar, view toggle, settings) is saved to it.
const colorMode = useColorMode();
const view = useView();
watch(me, (user) => {
  if (!user) return;
  colorMode.preference = user.theme;
  view.value = user.view as 'list' | 'grid';
});
watch([() => colorMode.preference, view], async ([theme, next]) => {
  if (!me.value || (theme === me.value.theme && next === me.value.view)) return;
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { theme, view: next } });
});

// Single keys, like Notion or Linear: Nuxt UI skips them while typing in a field or the editor.
defineShortcuts({
  '/': () => (searchOpen.value = true),
  n: () => navigateTo(newNotePath(currentFolder(route))),
  shift_n: () => createFolder(currentFolder(route)),
  'meta_\\': () => (collapsed.value = !collapsed.value),
});
</script>

<template>
  <!-- Nuxt UI's dashboard: a sidebar that resizes (its width kept in localStorage), collapses, and becomes a drawer on small screens. -->
  <UDashboardGroup unit="rem" storage="local" storage-key="notes-layout" class="gap-2 bg-muted p-2 text-default">
    <UDashboardSidebar
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

    <UDashboardPanel :ui="{ root: 'h-full min-h-0 flex-1 overflow-hidden rounded-xl border-0 bg-default', body: 'gap-0 p-0 sm:gap-0 sm:p-0' }">
      <template #body>
        <!-- Only on a small screen, where the sidebar is a drawer. -->
        <div class="flex px-2 pt-2 lg:hidden">
          <UDashboardSidebarToggle color="neutral" variant="ghost" />
        </div>
        <main class="flex min-h-0 flex-1 flex-col">
          <slot />
        </main>
      </template>
    </UDashboardPanel>

    <AppSearch />
  </UDashboardGroup>
</template>
