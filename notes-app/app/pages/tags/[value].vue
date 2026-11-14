<script setup lang="ts">
import type { TagDetail } from '~/types/items';

const { icons } = useIcons();

const route = useRoute();
const api = useApi();
const { renameTag, removeTag } = useItems();
const view = useView();

const value = route.params.value as string;
const {
  items: notes,
  status,
  error,
  load,
} = usePages(`tag:${value}`, async (query, onResponse) => {
  return (await api<TagDetail>(`/api${tagPath(value)}`, { query, onResponse, quiet: true })).notes;
});
useHead({ title: `#${value}` });
</script>

<template>
  <div class="p-8">
    <PageHeader
      :title="`#${value}`"
      :subtitle="notes.length || status === 'success' ? $t('tag.count', notes.length) : undefined"
    >
      <ViewToggle />
      <UTooltip :text="$t('common.rename')">
        <UButton
          :icon="icons.rename"
          color="neutral"
          variant="ghost"
          :aria-label="$t('tag.renameLabel')"
          @click="renameTag(value)"
        />
      </UTooltip>
      <UTooltip :text="$t('tag.deleteTooltip')">
        <UButton
          :icon="icons.delete"
          color="neutral"
          variant="ghost"
          :aria-label="$t('tag.deleteLabel')"
          @click="removeTag(value)"
        />
      </UTooltip>
    </PageHeader>
    <NoteList v-if="notes.length" :notes="notes" :grid="view === 'grid'" :load="load" />
    <UEmpty
      v-else-if="error"
      :icon="icons.tag"
      :title="$t('tag.notFound')"
      :description="$t('tag.notFoundDescription')"
      :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
    />
    <UEmpty
      v-else-if="status === 'success'"
      :icon="icons.tag"
      :title="$t('tag.empty')"
      :description="$t('tag.emptyDescription', { value })"
    />
  </div>
</template>
