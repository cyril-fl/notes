<script setup lang="ts">
const store = useDataStore();
const { hasLoaded } = storeToRefs(store);
const { handleDelete } = useDataApi();
</script>

<template>
  <main class="p-12">
    <div v-if="!hasLoaded">Loading</div>
    <template v-else>
      <header>
        <h1>My Notes App</h1>
        <button
          class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
          @click="handleDelete"
        >
          FLUSH DB
        </button>
      </header>

      <Navigation />

      <section>
        <slot />
      </section>

      <footer />
    </template>
  </main>
</template>
