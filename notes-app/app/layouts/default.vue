<script setup lang="ts">
const route = useRoute();
const { createFolder } = useItems();
const searchOpen = useState('search-open', () => false);

// Single keys, like Notion or Linear: Nuxt UI skips them while typing in a field or the editor.
defineShortcuts({
  '/': () => (searchOpen.value = true),
  n: () => navigateTo(newNotePath(currentFolder(route))),
  shift_n: () => createFolder(currentFolder(route)),
});
</script>

<template>
  <div class="flex h-dvh gap-2 overflow-hidden bg-muted p-2 text-default">
    <aside class="w-64 shrink-0 overflow-hidden rounded-xl bg-default">
      <AppSidebar />
    </aside>
    <main class="min-w-0 flex-1 overflow-y-auto rounded-xl bg-default">
      <slot />
    </main>
    <AppSearch />
  </div>
</template>
