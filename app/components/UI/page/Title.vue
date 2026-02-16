<script setup lang="ts">
import type { Descriptible, Titled } from '~/types/data';

// Define
export type PageTitleProps = Titled & Partial<Descriptible>;

// Data
defineProps<PageTitleProps>();
// Methods

// Lifecycle

// SEO
</script>

<template>
  <hgroup class="flex items-baseline gap-2">
    <h1 class="font-bold text-2xl">
      <slot name="title">
        {{ title }}
      </slot>
    </h1>
    <h3 v-if="description" class="text-muted text-sm">
      <slot name="description">
        {{ description }}
      </slot>
    </h3>
  </hgroup>
</template>

<style scoped></style>
