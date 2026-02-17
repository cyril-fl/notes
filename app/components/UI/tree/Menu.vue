<script setup lang="ts">
// Imports
import type { NavigationTreeProps } from '~/types/ui';
// Define
defineProps<NavigationTreeProps>();
// Data

// Methods

// Lifecycle

// SEO
</script>

<template>
  <!--  Group -->
  <UITreeMenuGroup v-if="item.children" :item="item">
    <UITreeMenu
      v-for="child in item.children"
      :key="`child-${child.id}`"
      :item="child"
    />
  </UITreeMenuGroup>

  <!--  Simple item -->
  <UITreeMenuItem v-else :item="item" />
</template>

<style scoped></style>
