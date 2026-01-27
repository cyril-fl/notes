<script setup lang="ts">
const store = useDataStore();
const { hasLoaded } = storeToRefs(store);

const { data, pending } = useAsyncData<ServerResponse<DataSchema[]>>(
  'elements',
  () => $fetch<ServerResponse<DataSchema[]>>('/api/data/read')
);

watch(
  () => data.value,
  (newValue) => {
    if (!newValue) return;
    store.onRead(newValue.data);
  },
  { immediate: true }
);

// ---- Test ----
const { handleDelete } = useDataApi();

</script>

<template>
  <main class="p-12">
    <div v-if="pending && !hasLoaded">Loading</div>
    <template v-else>
      <header>
        <h1>My Notes App</h1>
        <button class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted" @click="handleDelete">FLUSH DB</button>
      </header>

      <Navigation />

      <section>
        <slot />
      </section>

      <footer />
    </template>
  </main>
</template>
