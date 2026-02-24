<script setup lang="ts">
import type { NavigationTreeProps } from '~/types/ui';

defineProps<NavigationTreeProps>();
</script>

<template>
  <!--  Group -->
  <UINavigationTreeMenuGroup v-if="item.children" :item="item">
    <UINavigationTreeMenuNode
      v-for="child in item.children"
      :key="`child-${child.id}`"
      :item="child"
    />
  </UINavigationTreeMenuGroup>

  <!--  Simple item -->
  <UINavigationTreeMenuItem v-else :item="item" />
</template>

<style scoped></style>
