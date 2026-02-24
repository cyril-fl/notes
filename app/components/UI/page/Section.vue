<script setup lang="ts">
import type { PageTitleProps } from './Title.vue';
import type { Contextualizable } from '~/types/data';

type SectionProps = Partial<PageTitleProps & Contextualizable>;
const props = defineProps<SectionProps>();

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
      class="space-y-8 grow flex flex-col bg-default rounded-2xl p-4"
      v-bind="$attrs"
    >
      <slot v-if="!slots.header" name="header">
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

      <slot />
    </section>
  </UContextMenu>
</template>

<style scoped></style>
