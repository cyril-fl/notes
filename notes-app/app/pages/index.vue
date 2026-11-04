<script setup lang="ts">
const { data: items } = useChildren();
</script>

<template>
  <FolderView title="Notes" :items="items" />
</template>
