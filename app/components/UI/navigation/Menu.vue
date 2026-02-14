<script lang="ts" setup>
// Imports
import type { NavigationMenuProps } from '~/types/ui';
// Define
defineProps<NavigationMenuProps>();
// Data
// Methods
// Lifecycle
// SEO
</script>

<template>
  <!--  Group-->
  <UINavigationMenuGroup v-if="item.children" :item="item">
    <UINavigationMenu
      v-for="(child, index) in item.children"
      :key="`child-${child.label}-${index}`"
      :item="child"
    />
  </UINavigationMenuGroup>

  <!--  Simple item -->
  <UINavigationMenuItem v-else :item="item" />
</template>
