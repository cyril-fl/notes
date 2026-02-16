<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';

// Define

// Data
const route = useRoute();
const { t } = useI18n();
const icons = useIcons();
const { addFolder, addNote } = useActions();
const { $hooks } = useNuxtApp();

const id = computed<string>(() => {
  const pathParams = route.params.id;

  if (!pathParams || Array.isArray(pathParams)) return '';

  return pathParams;
});

const { getById } = useDataUtils();

const item = computed(() => getById(id.value, { types: ItemType.FOLDER }));

const children = computed(() => {
  const raw = item.value?.childrenIds;
  const result = {
    notes: [],
    folders: [],
  };

  if (!raw) return result;

  return raw.reduce<StoredData>((acc, id) => {
    const item = getById(id);
    if (!item) return acc;

    const callback = {
      [ItemType.NOTE]: () => {
        acc.notes.push(item as Note);
      },
      [ItemType.FOLDER]: () => {
        acc.folders.push(item as Folder);
      },
    };

    callback[item.type]();

    return acc;
  }, result);
});

const details = computed(() => {
  const _folder = t('pages.folder.folders_count', {
    count: children.value.folders.length,
  });
  const _note = t('pages.folder.notes_count', {
    count: children.value.notes.length,
  });

  const _result = [];

  if (children.value.folders.length) _result.push(_folder);
  if (children.value.notes.length) _result.push(_note);

  return _result.join(' | ');
});

const actions = ref<ContextMenuItem[][]>([
  [
    addFolder(id.value, (newId) => {
      $hooks.callHook('folder-card:update', newId);
    }),
    addNote(id.value),
  ],
]);

// Methods

// Lifecycle

// SEO
</script>

<template>
  <UIPageSection
    v-if="item"
    :title="item.title"
    :description="details"
    :context-actions="actions"
  >
    <UEmpty
      v-if="!children.folders.length && !children.notes.length"
      :icon="icons.folderempty"
      :title="t('pages.folder.empty')"
      class="grow"
      variant="naked"
    />
    <ul v-else class="flex flex-wrap gap-4">
      <li v-for="folder in children.folders" :key="folder.id">
        <UIFolderCard :item="folder" />
      </li>
      <li v-for="note in children.notes" :key="note.id">
        <UINotesCard :item="note" />
      </li>
    </ul>
  </UIPageSection>
  <UIPageSection v-else>
    <!-- <p>Item not found</p> -->
    <!-- <NuxtLink to="/notes"> Retour </NuxtLink> -->
    <UEmpty
      :icon="icons.folderempty"
      :title="t('pages.folder.not_found')"
      class="size-full"
      variant="naked"
    />
  </UIPageSection>
</template>

<style scoped></style>
