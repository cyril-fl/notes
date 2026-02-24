<script setup lang="ts">
import type { Descriptible, Titled } from '~/types/data';
import {
  EditableArea,
  EditableInput,
  EditablePreview,
  EditableRoot,
} from 'reka-ui';

export interface PageTitleProps extends Titled, Partial<Descriptible> {
  editable?: boolean;
}

const props = defineProps<PageTitleProps>();

const emit = defineEmits<{
  submit: [value: string];
  cancel: [];
}>();

const model = ref(props.title);
const editableRef = useTemplateRef('editableRef');

function handleSubmit(value: string | null | undefined) {
  const trimmed = value?.trim();
  if (trimmed && trimmed !== props.title) {
    emit('submit', trimmed);
  }
  emit('cancel');
}

function handleStateChange(state: 'edit' | 'submit' | 'cancel') {
  if (state === 'cancel') emit('cancel');
}

watch(
  () => props.title,
  (newTitle) => {
    model.value = newTitle;
  }
);

watch(
  () => props.editable,
  async (isEditable) => {
    if (!isEditable) return;
    model.value = props.title;
    await nextTick();
    editableRef.value?.edit();
  },
  { immediate: true }
);
</script>

<template>
  <hgroup class="flex items-baseline gap-2">
    <h1 class="font-bold text-2xl">
      <EditableRoot
        ref="editableRef"
        v-model="model"
        activation-mode="none"
        select-on-focus
        submit-mode="both"
        @submit="handleSubmit"
        @update:state="handleStateChange"
      >
        <EditableArea>
          <EditablePreview />
          <EditableInput
            class="placeholder-muted border-0 px-1 py-0.5 text-2xl bg-accented ring-transparent outline-none shadow-none rounded-sm"
          />
        </EditableArea>
      </EditableRoot>
    </h1>
    <h3 v-if="description && !editable" class="text-muted text-sm">
      <slot name="description">
        {{ description }}
      </slot>
    </h3>
  </hgroup>
</template>
