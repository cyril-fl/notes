<script setup lang="ts">
const { links } = useNavigation();
</script>

<template>
  <UINavigationLinkMenuNode
    v-for="(link, index) in links"
    :key="`link-${link.label}-${index}`"
    :item="link"
  />
</template>

<style scoped></style>
