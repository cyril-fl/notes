<script setup lang="ts">
import type { PageTitleProps } from './Title.vue';
import type { Contextualizable } from '~/types/data';

// Define
type SectionProps = Partial<PageTitleProps & Contextualizable>;
const props = defineProps<SectionProps>();
const slots = useSlots();

// Data
const isContextMenuActive = computed(
  () => props.contextActions && props.contextActions.length > 0
);
// Methods

// Lifecycle

// SEO
</script>

<template>
  <UContextMenu :items="props.contextActions" :disabled="!isContextMenuActive">
    <section class="space-y-8 grow flex flex-col" v-bind="$attrs">
      <!-- :disabled="!isContextMenuActive" -->
      <slot v-if="!slots.header" name="header">
        <header class="">
          <UIPageTitle
            v-if="props.title"
            :title="props.title"
            :description="props.description"
          />
        </header>
      </slot>

      <slot />
    </section>
  </UContextMenu>
</template>

<style scoped></style>
