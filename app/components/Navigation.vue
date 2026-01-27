<script setup lang="ts">
// Define
const store = useDataStore();
const { data, map } = storeToRefs(store);

// Data

// Methods

// Lifecycle

// SEO
</script>

<template>
  <nav>
    <menu>
      <li><NuxtLink to="/">Home</NuxtLink></li>
      <li><NuxtLink to="/notes">Notes</NuxtLink></li>
      <li><NuxtLink to="/notes/new">New Note</NuxtLink></li>
      <li v-for="folder in data.folders" :key="folder.id">
        <NuxtLink :to="folder.path.join('/')">
          {{ map.get(folder.id)?.title || $t('placeholder.folder_untitled') }}
        </NuxtLink>
      </li>
    </menu>
  </nav>
</template>
