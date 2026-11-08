<script setup lang="ts">
import type { ImportResult } from '~/types/items';

const emit = defineEmits<{ close: [] }>();

const { t } = useI18n();
const { icons } = useIcons();
const api = useApi();
const toast = useToast();
const destinations = useDestinations();

const parent = ref(ROOT);
const files = ref<File[]>([]);
const importing = ref(false);
const folderInput = useTemplateRef('folderInput');

// The files the API reads: a picked folder also holds what the import would skip (PDFs, .obsidian…),
// and its zips, which the API skips among other files and which would gray out the button.
const accept = '.zip,.md,image/*';
const kept = (file: File) => /\.md$/i.test(file.name) || file.type.startsWith('image/');

function addFolder(event: Event) {
  const input = event.target as HTMLInputElement;
  files.value = [...files.value, ...[...(input.files ?? [])].filter(kept)];
  input.value = ''; // The same folder can be picked again.
}

// A lone zip goes as it is; anything else as a form, each file named by its path, the folders made from it.
// A zip among other files would be skipped by the API: it goes alone.
const isZip = (file: File) => file.name.toLowerCase().endsWith('.zip');
const mixed = computed(() => files.value.length > 1 && files.value.some(isZip));

// Loose files are named by their name alone: two of the same name would be one note, the API refuses them.
const path = (file: File) => file.webkitRelativePath || file.name;
const duplicate = computed(() => files.value.map(path).find((name, i, names) => names.indexOf(name) !== i));

function body() {
  const [first] = files.value;
  if (files.value.length === 1 && first && isZip(first)) {
    return { body: first, headers: { 'Content-Type': 'application/zip' } };
  }
  const form = new FormData();
  for (const file of files.value) form.append('files', file, path(file));
  return { body: form };
}

async function submit() {
  importing.value = true;
  try {
    const result = await api<ImportResult>('/api/import', {
      method: 'POST',
      query: parent.value === ROOT ? {} : { parent: parent.value },
      ...body(),
      conflict: t('settings.importConflict'),
      tooLarge: t('settings.importTooLarge'),
    });
    toast.add({ title: t('settings.imported', { notes: result.notes, folders: result.folders }), color: 'success' });
    await refreshNuxtData();
    emit('close');
  } finally {
    importing.value = false;
  }
}
</script>

<template>
  <UModal :title="$t('settings.import')" :close="false" @update:open="emit('close')">
    <template #body>
      <form id="import" class="flex flex-col gap-4" @submit.prevent="submit">
        <UFormField :label="$t('settings.importInto')">
          <USelectMenu v-model="parent" :items="destinations" value-key="value" class="w-full" />
        </UFormField>
        <UFileUpload
          v-model="files"
          multiple
          layout="list"
          :accept="accept"
          :icon="icons.import"
          :label="$t('settings.importFiles')"
          :description="$t('settings.importFilesDescription')"
          :file-image="false"
          :ui="{ files: 'max-h-48 overflow-y-auto' }"
        >
          <template #actions>
            <!-- .stop: a click in the zone opens the file picker. -->
            <UButton
              :label="$t('settings.importFolder')"
              :icon="icons.folder"
              color="neutral"
              variant="outline"
              size="sm"
              @click.stop="folderInput?.click()"
            />
          </template>
        </UFileUpload>
        <input ref="folderInput" type="file" webkitdirectory class="hidden" @change="addFolder" />
        <p v-if="mixed" class="text-sm text-error">{{ $t('settings.importZipAlone') }}</p>
        <p v-if="duplicate" class="text-sm text-error">{{ $t('settings.importDuplicate', { name: duplicate }) }}</p>
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton
          type="submit"
          form="import"
          :label="$t('settings.import')"
          :loading="importing"
          :disabled="!files.length || mixed || !!duplicate"
        />
      </div>
    </template>
  </UModal>
</template>
