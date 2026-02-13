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
  <li>
    <UButton
      :icon="item.icon"
      :label="$t(item.label)"
      :to="item.to"
      class="w-full justify-start"
      color="neutral"
      variant="ghost"
    />
  </li>
</template>
