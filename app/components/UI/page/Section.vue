<script setup lang="ts">
import type { PageTitleProps } from './Title.vue';
import type { Contextualizable } from '~/types/data';

interface SectionProps extends Partial<PageTitleProps & Contextualizable> {
  searchable?: boolean;
  ancestors?: string[];
}
const props = withDefaults(defineProps<SectionProps>(), {
  searchable: false,
});

const emit = defineEmits<{
  submit: [value: string];
  cancel: [];
}>();

const slots = useSlots();

const isContextMenuActive = computed(
  () => props.contextActions && props.contextActions.length > 0
);

const hasBreadcrumb = computed(
  () =>
    props.ancestors &&
    props.ancestors.filter((id) => id !== 'root').length > 0
);
</script>

<template>
  <UContextMenu :items="props.contextActions" :disabled="!isContextMenuActive">
    <section
      class="space-y-4 grow flex flex-col bg-default rounded-2xl p-4"
      v-bind="$attrs"
    >
      <UICoreToolbar
        v-if="slots.header || props.title || searchable"
        as="header"
      >
        <template #left>
          <slot name="header">
            <UIPageTitle
              v-if="props.title"
              :title="props.title"
              :description="props.description"
              :editable="props.editable"
              @submit="emit('submit', $event)"
              @cancel="emit('cancel')"
            />
          </slot>
        </template>
        <template #right>
          <UIPageSearchbar v-if="searchable" />
        </template>
      </UICoreToolbar>

      <slot />

      <UICoreToolbar v-if="hasBreadcrumb" position="footer" class="mt-auto">
        <template #left>
          <UIPageBreadcrumb :ancestors="props.ancestors!" />
        </template>
      </UICoreToolbar>
    </section>
  </UContextMenu>
</template>

<style scoped></style>
