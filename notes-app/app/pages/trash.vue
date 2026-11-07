<script setup lang="ts">
import type { Trashed } from '~/types/items';

const { icons } = useIcons();
const { t, locale } = useI18n();

const api = useApi();
const { restoreItem, emptyTrash, purgeItem } = useItems();
const { confirm } = useDialogs();

const { data: items, status } = useLazyAsyncData('trash', () => api<Trashed[]>('/api/trash'), {
  default: () => [] as Trashed[],
});

async function askEmpty() {
  const confirmed = await confirm({
    title: t('trash.emptyConfirm'),
    description: t('trash.emptyConfirmDescription'),
    action: t('trash.emptyAction'),
  });
  if (confirmed) await emptyTrash();
}

async function askPurge(item: Trashed) {
  const confirmed = await confirm({
    title: t('trash.purgeConfirm', { title: displayTitle(item.title, t('untitled')) }),
    description: item.inside ? t('trash.purgeInside', { n: item.inside }) : t('trash.purgeAlone'),
    action: t('common.delete'),
  });
  if (confirmed) await purgeItem(item.id);
}

function deletedOn(item: Trashed) {
  return new Date(item.deleted_at).toLocaleDateString(locale.value, { day: 'numeric', month: 'short', year: 'numeric' });
}
</script>

<template>
  <div class="p-8">
    <PageHeader :title="$t('common.trash')">
      <UButton v-if="items.length" :label="$t('trash.emptyAction')" color="error" variant="ghost" @click="askEmpty" />
    </PageHeader>
    <ul v-if="items.length">
      <li v-for="item in items" :key="item.id" class="row h-10 border-b border-muted last:border-0">
        <UIcon :name="item.kind === 'folder' ? icons.folder : icons.note" class="size-4 shrink-0 text-muted" />
        <span class="truncate">{{ displayTitle(item.title, $t('untitled')) }}</span>
        <span v-if="item.inside" class="shrink-0 text-xs text-muted">+{{ item.inside }}</span>
        <span class="ml-auto shrink-0 pl-4 text-xs text-muted">{{ deletedOn(item) }}</span>
        <UButton :label="$t('trash.restore')" color="neutral" variant="ghost" size="xs" @click="restoreItem(item.id)" />
        <UTooltip :text="$t('trash.purge')">
          <UButton :icon="icons.close" color="neutral" variant="ghost" size="xs" :aria-label="$t('trash.purge')" @click="askPurge(item)" />
        </UTooltip>
      </li>
    </ul>
    <UEmpty
      v-else-if="status !== 'pending'"
      :icon="icons.trash"
      :title="$t('trash.empty')"
      :description="$t('trash.emptyDescription')"
    />
  </div>
</template>
