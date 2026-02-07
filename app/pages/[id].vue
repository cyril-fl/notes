<script setup lang="ts">
// Define

// Data
const route = useRoute();
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

// Methods

// Lifecycle

// SEO
</script>

<template>
  <p>Notes id</p>
  <div v-if="item" class="space-y-4">
    <div class="bg-muted p-4 rounded-md w-full text-xs text-muted">
      <pre>{{ item }}</pre>
    </div>

    <h1>{{ item.title }}</h1>

    <h3>Folders : {{ children.folders.length }}</h3>
    <ul>
      <li v-for="folder in children.folders" :key="folder.id">
        <NuxtLink :to="`/${folder.id}`" as="div">
          {{ folder.title }}
        </NuxtLink>
      </li>
    </ul>

    <h3>Notes : {{ children.notes.length }}</h3>
    <ul>
      <li v-for="note in children.notes" :key="note.id">
        <NuxtLink :to="`/notes/${note.id}`" as="div">
          {{ note.title }}
        </NuxtLink>
      </li>
    </ul>
  </div>
  <div v-else>
    <p>Item not found</p>
    <NuxtLink to="/notes"> Retour </NuxtLink>
  </div>
</template>

<style scoped></style>
