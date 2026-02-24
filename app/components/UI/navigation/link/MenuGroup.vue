<script lang="ts" setup>
import type { NavigationMenuProps } from '~/types/ui';

defineProps<NavigationMenuProps>();
</script>

<template>
  <ul class="space-y-1">
    <li
      v-if="!item.to"
      class="text-sm font-semibold text-muted px-2 select-none"
    >
      {{ $t(item.label) }}
    </li>
    <UINavigationLinkMenuItem v-else :item="item" />
    <slot />
  </ul>
</template>
