<script setup lang="ts">
import type { TagDetail } from '~/types/items';

const route = useRoute();
const api = useApi();
const { prompt } = useDialogs();
const view = useView();

const value = route.params.value as string;
const url = `/api/tags/${encodeURIComponent(value)}`;
const { data: tag, error } = useLazyAsyncData(`tag:${value}`, () => api<TagDetail>(url, { quiet: true }));

async function rename() {
  const next = await prompt({ title: 'Renommer le tag', value, action: 'Renommer' });
  if (!next || next === value) return;
  try {
    await api(url, { method: 'PATCH', body: { value: next }, conflict: 'Ce tag existe déjà.' });
    await navigateTo(`/tags/${encodeURIComponent(next)}`, { replace: true });
    await refreshNuxtData();
  } catch {
    // The toast already told the user.
  }
}

async function remove() {
  try {
    await api(url, { method: 'DELETE' });
    await navigateTo('/');
    await refreshNuxtData();
  } catch {
    // The toast already told the user.
  }
}
</script>

<template>
  <div class="p-8">
    <PageHeader :title="`#${value}`" :subtitle="tag ? `${tag.notes.length} note${tag.notes.length > 1 ? 's' : ''}` : undefined">
      <ViewToggle />
      <UTooltip text="Renommer">
        <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" aria-label="Renommer le tag" @click="rename" />
      </UTooltip>
      <UTooltip text="Supprimer le tag et l'effacer du texte de ses notes">
        <UButton icon="i-lucide-trash-2" color="neutral" variant="ghost" aria-label="Supprimer le tag" @click="remove" />
      </UTooltip>
    </PageHeader>
    <NoteList v-if="tag?.notes.length" :notes="tag.notes" :grid="view === 'grid'" />
    <UEmpty v-else-if="error" icon="i-lucide-hash" title="Ce tag n'existe pas" description="Il a peut-être été renommé ou supprimé." :actions="[{ label: 'Retour aux notes', icon: 'i-lucide-arrow-left', color: 'neutral', variant: 'subtle', to: '/' }]" />
    <UEmpty v-else-if="tag" icon="i-lucide-hash" title="Aucune note avec ce tag" :description="`Écrivez #${value} dans une note pour l'y retrouver.`" />
  </div>
</template>
