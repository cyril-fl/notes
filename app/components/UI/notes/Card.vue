<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';
import UIDndGhost from '~/components/UI/dnd/Ghost.vue';

interface NotesCardProps {
  item: Note;
}

const props = defineProps<NotesCardProps>();
const { deleteById, update } = useDataActions();
const { requestConfirm } = useConfirmDelete();
const { t } = useI18n();
const { icons } = useIcons();
const { isSelected, select, toggleSelect } = useSelection();

const cardRef = useTemplateRef('cardRef');
const itemRef = computed<Data>(() => props.item);

const { isDragging } = useDraggable({
  elementRef: cardRef,
  item: itemRef,
  ghostComponent: UIDndGhost,
});

const selected = computed(() => isSelected(props.item.id));

function handleClick(e: MouseEvent) {
  if (e.metaKey || e.ctrlKey) {
    toggleSelect(props.item.id);
  } else {
    select(props.item.id);
  }
}

function handleDblClick() {
  navigateTo(`${NAVIGATION.editor}${props.item.id}`);
}

const actions = computed<ContextMenuItem[][]>(() => [
  [
    {
      label: t('menu.context.delete'),
      icon: icons.delete,
      color: 'error' as const,
      onSelect: async () => {
        const confirmed = await requestConfirm(props.item.id, 'note');
        if (confirmed) deleteById(props.item.id);
      },
    },
  ],
  [
    {
      label: t(props.item.isPinned ? 'menu.context.unpin' : 'menu.context.pin'),
      icon: icons.pin,
      onSelect: () => {
        update(props.item.id, { isPinned: !props.item.isPinned });
      },
    },
    {
      label: t(props.item.isReadonly ? 'menu.context.unlock' : 'menu.context.lock'),
      icon: icons.lock,
      onSelect: () => {
        update(props.item.id, { isReadonly: !props.item.isReadonly });
      },
    },
  ],
]);
</script>

<template>
  <UContextMenu :items="actions" class="flex flex-col gap-2">
    <div
      ref="cardRef"
      class="relative bg-muted text-xs text-default p-4 rounded-md aspect-square size-30 overflow-hidden cursor-pointer select-none transition-all"
      :class="{
        'ring-2 ring-primary/50': selected,
        'opacity-40': isDragging,
      }"
      @click="handleClick"
      @dblclick="handleDblClick"
    >
      <span v-if="item.isPinned || item.isReadonly" class="absolute top-1 right-1 flex gap-1 text-dimmed">
        <UIcon v-if="item.isPinned" :name="icons.pin" class="size-3.5" />
        <UIcon v-if="item.isReadonly" :name="icons.lock" class="size-3.5" />
      </span>
      <span class="line-clamp-5">
        {{ item.content }}
      </span>
    </div>
    <ul class="text-xs text-muted text-center">
      <li>{{ item.updatedAt }}</li>
    </ul>
  </UContextMenu>
</template>
