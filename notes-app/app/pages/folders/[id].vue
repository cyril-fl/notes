<script setup lang="ts">
import type { ItemDetail } from '~/types/items';

const { icons } = useIcons();

const route = useRoute();
const api = useApi();

const id = route.params.id as string;
const { data: folder, error } = useLazyAsyncData(`folder:${id}`, () =>
  api<ItemDetail>(`/api/folders/${id}`, { quiet: true }),
);
</script>

<template>
  <FolderView
    v-if="folder"
    :folder-id="id"
    :title="folder.item.title"
    :subtitle="folder.path"
    :items="folder.children ?? []"
  />
  <UEmpty
    v-else-if="error"
    :icon="icons.folderNotFound"
    :title="$t('folder.notFound')"
    :description="$t('folder.notFoundDescription')"
    :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
  />
</template>
