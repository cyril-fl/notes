<script setup lang="ts">
import type { PageTitleProps } from './Title.vue';
import type { Contextualizable } from '~/types/data';

interface SectionProps extends Partial<PageTitleProps & Contextualizable> {
  searchable?: boolean;
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
</script>

<template>
  <UContextMenu :items="props.contextActions" :disabled="!isContextMenuActive">
    <section
      class="space-y-4 grow flex flex-col bg-default rounded-2xl p-4"
      v-bind="$attrs"
    >
      <UICoreToolbar v-if="slots.header || props.title || searchable">
        <template #left>
          <slot name="header">
            <header v-if="props.title">
              <UIPageTitle
                :title="props.title"
                :description="props.description"
                :editable="props.editable"
                @submit="emit('submit', $event)"
                @cancel="emit('cancel')"
              />
            </header>
          </slot>
        </template>
        <template #right>
          <UIPageSearchbar v-if="searchable" />
        </template>
      </UICoreToolbar>

      <slot />
    </section>
  </UContextMenu>
</template>

<style scoped></style>
