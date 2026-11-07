<script setup lang="ts">
import type { TagDetail } from '~/types/items';

const { icons } = useIcons();

const route = useRoute();
const api = useApi();
const { renameTag, removeTag } = useItems();
const view = useView();

const value = route.params.value as string;
const { data: tag, error } = useLazyAsyncData(`tag:${value}`, () => api<TagDetail>(`/api${tagPath(value)}`, { quiet: true }));
</script>

<template>
  <div class="p-8">
    <PageHeader :title="`#${value}`" :subtitle="tag ? $t('tag.count', tag.notes.length) : undefined">
      <ViewToggle />
      <UTooltip :text="$t('common.rename')">
        <UButton :icon="icons.rename" color="neutral" variant="ghost" :aria-label="$t('tag.renameLabel')" @click="renameTag(value)" />
      </UTooltip>
      <UTooltip :text="$t('tag.deleteTooltip')">
        <UButton :icon="icons.delete" color="neutral" variant="ghost" :aria-label="$t('tag.deleteLabel')" @click="removeTag(value)" />
      </UTooltip>
    </PageHeader>
    <NoteList v-if="tag?.notes.length" :notes="tag.notes" :grid="view === 'grid'" />
    <UEmpty v-else-if="error" :icon="icons.tag" :title="$t('tag.notFound')" :description="$t('tag.notFoundDescription')" :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]" />
    <UEmpty v-else-if="tag" :icon="icons.tag" :title="$t('tag.empty')" :description="$t('tag.emptyDescription', { value })" />
  </div>
</template>
