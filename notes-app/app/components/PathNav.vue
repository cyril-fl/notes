<script setup lang="ts">
/** The folder path, '/A/B', as a breadcrumb, with the history's back and forward buttons. */
const props = defineProps<{ path: string }>();

const { icons } = useIcons();
const { t } = useI18n();

const destinations = useDestinations();

// shortcut: matches folders by path, a folder name holding '/' gets no link; ask the API for the ancestors' ids if that matters.
const breadcrumb = computed(() => {
  const byPath = new Map(destinations.value.map((folder) => [folder.label, folder.value]));
  const names = props.path.split('/').filter(Boolean);
  return [
    { label: t('common.notes'), to: '/' },
    ...names.map((name, index) => {
      const folderId = byPath.get(`/${names.slice(0, index + 1).join('/')}`);
      return { label: name, to: folderId && `/folders/${folderId}` };
    }),
  ];
});

const router = useRouter();
const route = useRoute();
// vue-router keeps the neighbours of the current entry in history.state; route.fullPath makes it re-read on each navigation.
const canGoBack = computed(() => route.fullPath && !!window.history.state?.back);
const canGoForward = computed(() => route.fullPath && !!window.history.state?.forward);
</script>

<template>
  <!-- Plain muted text like the other pages' subtitles: the path reads as before, its parts just link. -->
  <nav class="flex min-w-0 items-center gap-1 text-xs text-muted" :aria-label="$t('path.breadcrumb')">
    <button
      type="button"
      class="shrink-0 hover:text-default disabled:opacity-40 disabled:hover:text-muted"
      :aria-label="$t('path.back')"
      :disabled="!canGoBack"
      @click="router.back()"
    >
      <UIcon :name="icons.back" class="block size-3.5" />
    </button>
    <button
      type="button"
      class="mr-1 shrink-0 hover:text-default disabled:opacity-40 disabled:hover:text-muted"
      :aria-label="$t('path.forward')"
      :disabled="!canGoForward"
      @click="router.forward()"
    >
      <UIcon :name="icons.forward" class="block size-3.5" />
    </button>
    <template v-for="(part, index) in breadcrumb" :key="index">
      <span v-if="index" aria-hidden="true">/</span>
      <NuxtLink v-if="part.to" :to="part.to" class="truncate hover:text-default">{{ part.label }}</NuxtLink>
      <span v-else class="truncate">{{ part.label }}</span>
    </template>
  </nav>
</template>
