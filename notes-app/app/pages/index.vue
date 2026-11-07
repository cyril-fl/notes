<script setup lang="ts">
const { data: items } = useChildren();
</script>

<template>
  <FolderView :title="$t('common.notes')" :items="items" />
</template>
