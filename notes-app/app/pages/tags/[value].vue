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
    <PageHeader :title="`#${value}`" :subtitle="tag ? `${tag.notes.length} note${tag.notes.length > 1 ? 's' : ''}` : undefined">
      <ViewToggle />
      <UTooltip text="Renommer">
        <UButton :icon="icons.rename" color="neutral" variant="ghost" aria-label="Renommer le tag" @click="renameTag(value)" />
      </UTooltip>
      <UTooltip text="Supprimer le tag et l'effacer du texte de ses notes">
        <UButton :icon="icons.delete" color="neutral" variant="ghost" aria-label="Supprimer le tag" @click="removeTag(value)" />
      </UTooltip>
    </PageHeader>
    <NoteList v-if="tag?.notes.length" :notes="tag.notes" :grid="view === 'grid'" />
    <UEmpty v-else-if="error" :icon="icons.tag" title="Ce tag n'existe pas" description="Il a peut-être été renommé ou supprimé." :actions="[{ label: 'Retour aux notes', icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]" />
    <UEmpty v-else-if="tag" :icon="icons.tag" title="Aucune note avec ce tag" :description="`Écrivez #${value} dans une note pour l'y retrouver.`" />
  </div>
</template>
