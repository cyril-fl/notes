<script setup lang="ts">
const props = defineProps<{ parentId?: string }>();

const { data: items } = useChildren(props.parentId);
const folders = computed(() => items.value.filter((item) => item.kind === 'folder'));
</script>

<template>
  <ul v-if="folders.length" :class="{ 'ml-3': parentId }">
    <li v-for="folder in folders" :key="folder.id">
      <FolderNode :folder="folder" />
    </li>
  </ul>
</template>
