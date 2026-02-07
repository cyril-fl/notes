<script setup lang="ts">
const props = defineProps<{ folder: Folder }>();

const store = useDataStore();
const { data, map } = storeToRefs(store);
const { handleCreateFolderInFolder, handleDeleteId } = useDataApi();
const { getById } = useDataUtils();

const count = computed(() => data.value.folders.length);

const children = computed(() => {
  return props.folder.childrenIds.map((id) =>
    getById(id, { types: ItemType.FOLDER })
  );
});
</script>

<template>
  <li class="space-y-2 space-x-2">
    <NuxtLink
      :to="`/${folder.id}`"
      class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
    >
      {{ map.get(folder.id)?.title || $t('placeholder.folder_untitled') }}
    </NuxtLink>

    <button
      class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
      @click="
        () =>
          handleCreateFolderInFolder({
            folder,
            title: folder.title + ' - ' + count + 1,
          })
      "
    >
      Create Sub Folder
    </button>
    <button
      class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
      @click="handleDeleteId(folder.id)"
    >
      DELETE
    </button>
    <NuxtLink
      class="bg-inverted text-inverted p-1 rounded-sm hover:bg-muted hover:text-muted"
      :to="`/notes/new/${folder.id}`"
    >
      Create Note {{ folder.path.join('/') }}
    </NuxtLink>
    <ul v-if="children.length > 0">
      <template v-for="child in children">
        <NavigationFolder v-if="child" :key="child.id" :folder="child" />
      </template>
    </ul>
  </li>
</template>
