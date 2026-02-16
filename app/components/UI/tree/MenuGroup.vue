<script setup lang="ts">
// Imports
import type { NavigationTreeProps } from '~/types/ui';
// Define
defineProps<NavigationTreeProps>();
const { getListStateIcon } = useIcons();

// Data

// Methods

// Lifecycle

// SEO
</script>

<template>
  <div class="space-y-1">
    <p v-if="!item.to" class="text-sm font-semibold text-muted px-2">
      {{ $t(item.label) }}
    </p>
    <template v-else>
      <UCollapsible>
        <template #default="{ open }">
          <div class="flex items-center">
            <UITreeMenuItem :item="item" />
            <UIcon :name="getListStateIcon(open)" class="ml-2" />
          </div>
        </template>

        <template #content>
          <div class="space-y-1 pl-4">
            <slot />
          </div>
        </template>
      </UCollapsible>
    </template>
    <slot v-if="!item.to" />
  </div>
</template>

<style scoped></style>
