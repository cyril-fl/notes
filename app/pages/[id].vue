<script setup lang="ts">
// Define

// Data
const route = useRoute();
const { t } = useI18n();
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
  const _folder = t('pages.folder.folders_count', { count: children.value.folders.length });
  const _note = t('pages.folder.notes_count', { count: children.value.notes.length });

  const result = []

  if (children.value.folders.length) result.push(_folder);
  if (children.value.notes.length) result.push(_note);

  return result.join(' | ');
});

// Methods

// Lifecycle

// SEO
</script>

<template>
  <div v-if="item" class="space-y-4">
    <hgroup class="flex items-baseline  gap-2">
      <h1 class="font-bold text-2xl">{{ item.title }}</h1>
      <h3>{{ details }}</h3>
    </hgroup>

    <ul>
      <li v-for="folder in children.folders" :key="folder.id">
        <NuxtLink :to="`/${folder.id}`" as="div">
          {{ folder.title }}
        </NuxtLink>
      </li>
    </ul>

    <ul class="flex flew-wrap gap-4">
      <li v-for="note in children.notes" :key="note.id">
        <UINotesCard :note="note" />
      </li>
    </ul>
  </div>
  <div v-else>
    <p>Item not found</p>
    <NuxtLink to="/notes"> Retour </NuxtLink>
  </div>
</template>

<style scoped></style>
