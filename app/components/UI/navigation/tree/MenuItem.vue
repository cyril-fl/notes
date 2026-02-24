<script setup lang="ts">
import type { NavigationTreeProps } from '~/types/ui';
import { EditableArea, EditableInput, EditableRoot } from 'reka-ui';

const props = defineProps<NavigationTreeProps>();
const { set, editingId, clearEdit } = useCurrentFolder();
const { update } = useDataActions();

const isEditing = computed(() => editingId.value === props.item.id);
const model = ref(props.item.label);
const editable = useTemplateRef('editable');

function handleSubmit(value: string | null | undefined) {
  const trimmed = value?.trim();

  if (trimmed && trimmed !== props.item.label) {
    update(props.item.id, { title: trimmed });
  }

  clearEdit();
}

function handleStateChange(state: 'edit' | 'submit' | 'cancel') {
  if (state === 'cancel') clearEdit();
}

watch(
  () => props.item.label,
  (newLabel) => {
    model.value = newLabel;
  }
);

watch(isEditing, async (editing) => {
  if (!editing) return;
  model.value = props.item.label;
  await nextTick();
  editable.value?.edit();
});
</script>

<template>
  <li @mouseenter="set(props.item.id)" class="select-none">
    <UButton
      :icon="item.icon"
      :to="item.to"
      class="w-full justify-start"
      color="neutral"
      variant="ghost"
    >
      <EditableRoot
        ref="editable"
        v-model="model"
        activation-mode="none"
        select-on-focus
        submit-mode="both"
        auto-resize
        @submit="handleSubmit"
        @update:state="handleStateChange"
      >
        <EditableArea
          v-if="isEditing"
          class="min-w-0 flex-1"
          @click.stop.prevent
          @mousedown.stop
        >
          <EditableInput
            class="placeholder-muted border-0 px-1 py-0.5 text-sm bg-accented ring-transparent outline-none shadow-none rounded-sm"
          />
        </EditableArea>
        <span v-else class="truncate">{{ item.label }}</span>
      </EditableRoot>
    </UButton>
  </li>
</template>
