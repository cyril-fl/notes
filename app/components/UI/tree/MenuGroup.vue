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
  <ul class="space-y-1">
    <li v-if="!item.to" class="text-sm font-semibold text-muted px-2">
      {{ $t(item.label) }}
    </li>
    <UITreeMenuItem v-else :item="item" />
    <slot />
  </ul>
</template>

<style scoped></style>
