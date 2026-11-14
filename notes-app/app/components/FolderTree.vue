<script setup lang="ts">
const props = defineProps<{ parentId?: string }>();

// Folders only: a folder of thousands of notes would download them all for nothing.
const { data: folders } = useChildren(props.parentId, 'folder');
</script>

<template>
  <ul v-if="folders.length" :class="{ 'ml-3': parentId }">
    <li v-for="folder in folders" :key="folder.id">
      <FolderNode :folder="folder" />
    </li>
  </ul>
</template>
