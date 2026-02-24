<script lang="ts" setup>
import type { NavigationMenuProps } from '~/types/ui';

defineProps<NavigationMenuProps>();
</script>

<template>
  <!--  Group-->
  <UINavigationLinkMenuGroup v-if="item.children" :item="item">
    <UINavigationLinkMenuNode
      v-for="(child, index) in item.children"
      :key="`child-${child.id}-${index}`"
      :item="child"
    />
  </UINavigationLinkMenuGroup>

  <!--  Simple item -->
  <UINavigationLinkMenuItem v-else :item="item" />
</template>
