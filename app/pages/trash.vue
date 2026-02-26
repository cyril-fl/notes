<script setup lang="ts">
import type { ContextMenuItem } from '@nuxt/ui';

const store = useDataStore();
const { trash, trashCount } = storeToRefs(store);
const { restoreById, purgeById, emptyTrash } = useDataActions();
const { requestConfirm } = useConfirmDelete();
const { t } = useI18n();
const { icons } = useIcons();

const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;

function daysRemaining(deletedAt: Date | string): number {
  const deleted = new Date(deletedAt).getTime();
  const expiry = deleted + THIRTY_DAYS_MS;
  const remaining = Math.ceil((expiry - Date.now()) / (24 * 60 * 60 * 1000));
  return Math.max(0, remaining);
}

function getItemTitle(item: DataSchema): string {
  if (item.type === ItemType.FOLDER) return item.title;
  return extractTitleFromContent(item.content) || t('notes.default_title');
}

function getItemIcon(item: DataSchema): string {
  return item.type === ItemType.FOLDER ? icons.folder : icons.note;
}

function trashItemActions(item: DataSchema): ContextMenuItem[][] {
  return [
    [
      {
        label: t('menu.context.restore'),
        icon: icons.restore,
        onSelect: () => restoreById(item.id),
      },
      {
        label: t('menu.context.purge'),
        icon: icons.delete,
        color: 'error' as const,
        onSelect: async () => {
          const itemType =
            item.type === ItemType.FOLDER ? 'folder' : 'note';
          const confirmed = await requestConfirm(
            item.id,
            itemType as 'note' | 'folder'
          );
          if (confirmed) purgeById(item.id);
        },
      },
    ],
  ];
}

async function handleEmptyTrash() {
  const { requestConfirm: confirm } = useConfirmDelete();
  const confirmed = await confirm('trash', 'folder');
  if (confirmed) await emptyTrash();
}

const actions = computed<ContextMenuItem[][]>(() => [
  [
    {
      label: t('pages.trash.empty_action'),
      icon: icons.delete,
      color: 'error' as const,
      disabled: trashCount.value === 0,
      onSelect: handleEmptyTrash,
    },
  ],
]);

usePageSection(
  computed(() => ({
    title: t('pages.trash.title'),
    description: t('pages.folder.notes_count', {
      count: trashCount.value,
    }),
    contextActions: actions.value,
  }))
);
</script>

<template>
  <UEmpty
    v-if="!trash.length"
    :icon="icons.trash"
    :title="t('pages.trash.empty')"
    class="grow"
    variant="naked"
  />
  <ul v-else class="space-y-2">
    <li v-for="item in trash" :key="item.id">
      <UContextMenu :items="trashItemActions(item)">
        <div
          class="flex items-center gap-3 p-3 rounded-md bg-muted hover:bg-accented transition-colors"
        >
          <UIcon :name="getItemIcon(item)" class="size-5 text-muted shrink-0" />
          <div class="flex-1 min-w-0">
            <p class="text-sm text-default truncate">
              {{ getItemTitle(item) }}
            </p>
            <p class="text-xs text-muted">
              {{
                t('pages.trash.days_remaining', {
                  days: daysRemaining(item.deletedAt!),
                })
              }}
            </p>
          </div>
        </div>
      </UContextMenu>
    </li>
  </ul>
</template>
