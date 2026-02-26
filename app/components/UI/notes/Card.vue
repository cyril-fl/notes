<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';

interface NotesCardProps {
  item: Note;
}

const props = defineProps<NotesCardProps>();
const { deleteById } = useDataActions();
const { t } = useI18n();
const { icons } = useIcons();

const actions = computed<ContextMenuItem[][]>(() => [
  [
    {
      label: t('menu.context.delete'),
      icon: icons.delete,
      color: 'error' as const,
      onSelect: () => deleteById(props.item.id),
    },
  ],
]);
</script>

<template>
  <UContextMenu :items="actions" class="flex flex-col gap-2">
    <NuxtLink
      :to="`${NAVIGATION.editor}${item.id}`"
      as="p"
      class="bg-muted text-xs text-default p-4 rounded-md aspect-square size-30 overflow-hidden"
    >
      <span class="line-clamp-5">
        {{ item.content }}
      </span>
    </NuxtLink>
    <ul class="text-xs text-muted text-center">
      <li>{{ item.updatedAt }}</li>
    </ul>
  </UContextMenu>
</template>
