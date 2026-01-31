<script setup lang="ts">
// Define
const { data, tree } = storeToRefs(useDataStore());
const { handleCreate } = useDataApi();

const count = computed(() => data.value.folders.length);

const rootFolders = computed(() =>
  data.value.folders.filter((f) => f.path.length === 1)
);

// Methods
const handleCreateFolder = async () => {
  await handleCreate({
    path: [],
    type: ItemType.FOLDER,
    title: `New Folder ${count.value + 1}`,
    childrenIds: [],
  });
};
</script>

<template>
  <nav>
    <menu class="space-y-2">
      <li><NuxtLink to="/">Accueil</NuxtLink></li>
      <li><NuxtLink to="/notes/">List des notes</NuxtLink></li>
      <li><NuxtLink to="/notes/new">Nouvelle note</NuxtLink></li>
      <li>
        <button
          class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
          @click="handleCreateFolder"
        >
          Create Folder
        </button>
      </li>
      <li v-if="data.folders.length <= 0">No folder found</li>
      <template v-else>
        <NavigationFolder
          v-for="folder in rootFolders"
          :key="folder.id"
          :folder="folder"
        />
      </template>
    </menu>

    <pre>{{ tree }}</pre>
  </nav>
</template>
