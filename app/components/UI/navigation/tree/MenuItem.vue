<script setup lang="ts">
import type { NavigationTreeProps } from '~/types/ui';
import { EditableArea, EditableInput, EditableRoot } from 'reka-ui';
import UIDndGhost from '~/components/UI/dnd/Ghost.vue';

const props = defineProps<NavigationTreeProps>();
const { set, editingId, clearEdit } = useCurrentFolder();
const { update } = useDataActions();
const { getById } = useDataUtils();

const isEditing = computed(() => editingId.value === props.item.id);

const dropRef = useTemplateRef('dropRef');
const folderIdRef = computed(() => props.item.id);
const { isOver } = useDropZone({ elementRef: dropRef, folderId: folderIdRef });

// Make navbar folders draggable (blocked during editing)
const folderData = computed<Data>(() => getById(props.item.id)!);
const { isDragging } = useDraggable({
  elementRef: dropRef,
  item: folderData,
  ghostComponent: UIDndGhost,
  canDragCheck: () => !isEditing.value,
});
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
  <li ref="dropRef" @mouseenter="set(props.item.id)" class="select-none transition-colors" :class="{ 'bg-primary/10 rounded-md': isOver, 'opacity-40': isDragging }">
    <UButton
      :icon="item.icon"
      :to="item.to"
      draggable="false"
      class="w-full justify-start"
      color="neutral"
      variant="ghost"
      block
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
