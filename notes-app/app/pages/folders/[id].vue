<script setup lang="ts">
import type { ItemDetail } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

const route = useRoute();
const api = useApi();

const id = route.params.id as string;
const { data: folder, error } = useLazyAsyncData(`folder:${id}`, () =>
  api<ItemDetail>(`/api/folders/${id}`, { quiet: true }),
);
useHead({ title: () => folder.value && displayTitle(folder.value.item.title, t('untitled')) });
</script>

<template>
  <FolderView v-if="folder" :folder-id="id" :title="folder.item.title" :subtitle="folder.path" />
  <UEmpty
    v-else-if="error"
    :icon="icons.folderNotFound"
    :title="$t('folder.notFound')"
    :description="$t('folder.notFoundDescription')"
    :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
  />
</template>
