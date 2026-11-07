<script setup lang="ts">
import type { FolderPath, ImportResult } from '~/types/items';

const api = useApi();
const toast = useToast();
// The API sends the zip as an attachment: a plain link downloads it.
const exportURL = `${useRuntimeConfig().public.apiBase}/api/export`;
const input = useTemplateRef('input');
const importing = ref(false);

// The select refuses '' as a value: ROOT stands for the root, which the API takes as no parent.
const ROOT = 'root';
const parent = ref(ROOT);
const { data: folders } = useLazyAsyncData('folder-paths', () => api<FolderPath[]>('/api/folders'), { default: () => [] as FolderPath[] });
const destinations = computed(() => [{ label: 'Racine', value: ROOT }, ...folders.value.map((folder) => ({ label: folder.path, value: folder.id }))]);

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

    <section>
      <h2 class="mb-1 text-sm font-semibold">Données</h2>
      <ul>
        <li class="flex items-center gap-4 border-b border-muted py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">Exporter</p>
            <p class="text-xs text-muted">Toutes les notes en fichiers markdown, un dossier par dossier, dans un zip.</p>
          </div>
          <UButton label="Exporter" icon="i-lucide-download" color="neutral" variant="outline" :href="exportURL" download />
        </li>
        <li class="flex items-center gap-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">Importer</p>
            <p class="text-xs text-muted">Les fichiers .md d'un zip (un export, un vault Obsidian…), ajoutés au dossier choisi.</p>
          </div>
          <USelectMenu v-model="parent" :items="destinations" value-key="value" aria-label="Importer dans" class="w-56" />
          <input ref="input" type="file" accept=".zip,application/zip" class="hidden" @change="importZip">
          <UButton label="Importer" icon="i-lucide-upload" color="neutral" variant="outline" :loading="importing" @click="input?.click()" />
        </li>
      </ul>
    </section>
  </div>
</template>
