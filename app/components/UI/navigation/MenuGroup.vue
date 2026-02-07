<script lang="ts" setup>
import type { NavigationMenuProps } from '~/types/ui';

defineProps<NavigationMenuProps>();
</script>

<template>
  <ul class="space-y-1">
    <li v-if="!item.to" class="text-sm font-semibold text-muted px-2">
      {{ $t(item.label) }}
    </li>
    <UINavigationMenuItem v-else :item="item" />
    <slot />
  </ul>
</template>
