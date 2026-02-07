<script lang="ts" setup>
// Imports
import type { NavigationMenuItem } from '~/types/ui';

// Define
defineProps<{
  item: NavigationMenuItem;
}>();
// Data
// Methods
// Lifecycle
// SEO
</script>

<template>
  <li>
    <UIButton
      :icon="item.icon"
      :label="$t(item.label)"
      :to="item.to"
      class="w-full justify-start"
      color="neutral"
      variant="ghost"
    />
  </li>
</template>

<style scoped />
