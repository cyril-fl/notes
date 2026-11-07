<script setup lang="ts">
import type { ImportResult, User } from '~/types/items';

const { icons } = useIcons();

const api = useApi();
const toast = useToast();
// The API sends the zip as an attachment: a plain link downloads it.
const exportURL = `${useRuntimeConfig().public.apiBase}/api/export`;
const input = useTemplateRef('input');
const importing = ref(false);

const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));

async function logout() {
  await api('/api/logout', { method: 'POST' });
  clearNuxtData(); // The next account must not see what this one read.
  await navigateTo('/login');
}

const parent = ref(ROOT);
const destinations = useDestinations();

async function importZip(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) return;
  importing.value = true;
  try {
    const result = await api<ImportResult>('/api/import', {
      method: 'POST',
      query: parent.value === ROOT ? {} : { parent: parent.value },
      body: file,
      headers: { 'Content-Type': 'application/zip' },
      conflict: 'Un dossier du zip existe déjà à cet endroit : importez ailleurs, ou renommez-le.',
    });
    toast.add({ title: `${result.notes} note(s) et ${result.folders} dossier(s) importés`, color: 'success' });
    await refreshNuxtData();
  } finally {
    importing.value = false;
    input.value!.value = ''; // The same file can be picked again.
  }
}
</script>

<template>
  <div class="p-8">
    <PageHeader title="Réglages" />

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">Compte</h2>
      <div class="flex items-center gap-4 py-3">
        <p class="min-w-0 flex-1 truncate text-sm">
          Connecté en tant que <span class="font-medium">{{ me?.name }}</span>
        </p>
        <UButton label="Se déconnecter" color="neutral" variant="outline" @click="logout" />
      </div>
    </section>

    <section>
      <h2 class="mb-1 text-sm font-semibold">Données</h2>
      <ul>
        <li class="flex items-center gap-4 border-b border-muted py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">Exporter</p>
            <p class="text-xs text-muted">Toutes les notes en fichiers markdown, un dossier par dossier, dans un zip.</p>
          </div>
          <UButton label="Exporter" :icon="icons.export" color="neutral" variant="outline" :href="exportURL" download />
        </li>
        <li class="flex items-center gap-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">Importer</p>
            <p class="text-xs text-muted">Les fichiers .md d'un zip (un export, un vault Obsidian…), ajoutés au dossier choisi.</p>
          </div>
          <USelectMenu v-model="parent" :items="destinations" value-key="value" aria-label="Importer dans" class="w-56" />
          <input ref="input" type="file" accept=".zip,application/zip" class="hidden" @change="importZip">
          <UButton label="Importer" :icon="icons.import" color="neutral" variant="outline" :loading="importing" @click="input?.click()" />
        </li>
      </ul>
    </section>
  </div>
</template>
