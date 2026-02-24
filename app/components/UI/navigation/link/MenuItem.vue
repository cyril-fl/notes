<script lang="ts" setup>
import type { NavigationMenuItem } from '~/types/ui';

defineProps<{
  item: NavigationMenuItem;
}>();
</script>

<template>
  <li>
    <UButton
      :icon="item.icon"
      :label="$t(item.label)"
      :to="item.to"
      class="w-full justify-start select-none"
      color="neutral"
      variant="ghost"
    />
  </li>
</template>
