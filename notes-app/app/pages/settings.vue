<script setup lang="ts">
import type { ImportResult, Snippet, User } from '~/types/items';

const { icons } = useIcons();
const { t, locale } = useI18n();

const api = useApi();
const toast = useToast();
// The API sends the zip as an attachment: a plain link downloads it.
const exportURL = `${useRuntimeConfig().public.apiBase}/api/export`;
const input = useTemplateRef('input');
const importing = ref(false);

const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));
const avatarSrc = computed(
  () => me.value?.avatar && imageSrc(`/api/files/${me.value.avatar}`, useRuntimeConfig().public.apiBase),
);

// Each language in its own words. The layout watches me: it switches the interface once the API saved the choice.
const languages = [
  { label: 'Français', value: 'fr' },
  { label: 'English', value: 'en' },
];

async function changeLanguage(next: string) {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { locale: next } });
}

// The layout applies the theme of me, and saves the mode to the account whenever it changes.
const colorMode = useColorMode();
const modes = computed(() => [
  { label: t('settings.modeSystem'), value: 'system' },
  { label: t('settings.modeLight'), value: 'light' },
  { label: t('settings.modeDark'), value: 'dark' },
]);

async function changeTheme(next: string) {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { theme: next } });
}

// 0 never purges; a retention set by hand elsewhere stays in the list.
const retentions = computed(() =>
  [...new Set([0, 7, 30, 90, 365, me.value?.trash_retention_days ?? 30])]
    .sort((a, b) => a - b)
    .map((n) => ({ label: n ? t('settings.retentionDays', n) : t('settings.retentionNever'), value: n })),
);

async function changeRetention(next: number) {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { trash_retention_days: next } });
}

const { confirm, snippet: editSnippet } = useDialogs();

// The API takes the whole list: an edit sends it back with the one changed.
async function saveSnippets(snippets: Snippet[]) {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { snippets } });
}

async function addSnippet() {
  const added = await editSnippet({ title: t('settings.addSnippet') });
  if (added) await saveSnippets([...(me.value?.snippets ?? []), added]);
}

async function changeSnippet(index: number) {
  const snippets = [...(me.value?.snippets ?? [])];
  const changed = await editSnippet({ title: t('settings.editSnippet'), value: snippets[index] });
  if (!changed) return;
  snippets[index] = changed;
  await saveSnippets(snippets);
}

async function removeSnippet(index: number) {
  const snippets = [...(me.value?.snippets ?? [])];
  const confirmed = await confirm({
    title: t('settings.deleteSnippet', { name: snippets[index]?.name }),
    action: t('common.delete'),
  });
  if (!confirmed) return;
  snippets.splice(index, 1);
  await saveSnippets(snippets);
}

async function deleteData() {
  const confirmed = await confirm({
    title: t('settings.deleteData'),
    description: t('settings.deleteDataConfirm'),
    action: t('settings.deleteData'),
    expected: me.value?.name,
  });
  if (!confirmed) return;
  await api('/api/me/data', { method: 'DELETE' });
  toast.add({ title: t('settings.dataDeleted'), color: 'success' });
  await refreshNuxtData();
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
      conflict: t('settings.importConflict'),
    });
    toast.add({ title: t('settings.imported', { notes: result.notes, folders: result.folders }), color: 'success' });
    await refreshNuxtData();
  } finally {
    importing.value = false;
    input.value!.value = ''; // The same file can be picked again.
  }
}
</script>

<template>
  <div class="p-8">
    <PageHeader :title="$t('common.settings')" />

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.account') }}</h2>
      <div class="flex items-center gap-4 py-3">
        <UAvatar :src="avatarSrc || undefined" :alt="me?.name" />
        <p class="min-w-0 flex-1 truncate text-sm">
          {{ $t('settings.loggedInAs') }} <span class="font-medium">{{ me?.name }}</span>
        </p>
        <UButton
          :label="$t('settings.manageAccount')"
          :icon="icons.account"
          to="/account"
          color="neutral"
          variant="outline"
        />
      </div>
    </section>

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.language') }}</h2>
      <div class="flex items-center gap-4 py-3">
        <p class="min-w-0 flex-1 text-xs text-muted">{{ $t('settings.languageDescription') }}</p>
        <USelectMenu
          :model-value="locale"
          :items="languages"
          value-key="value"
          :search-input="false"
          :aria-label="$t('settings.language')"
          class="w-56"
          @update:model-value="changeLanguage"
        />
      </div>
    </section>

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.appearance') }}</h2>
      <ul>
        <li class="border-b border-muted py-3">
          <p class="text-sm">{{ $t('settings.theme') }}</p>
          <p class="mb-3 text-xs text-muted">{{ $t('settings.themeDescription') }}</p>
          <div class="grid grid-cols-[repeat(auto-fill,minmax(10rem,1fr))] gap-3">
            <ThemePreview
              v-for="theme in themes"
              :key="theme.id"
              :theme="theme"
              :active="me?.theme === theme.id"
              @click="changeTheme(theme.id)"
            />
          </div>
        </li>
        <li class="flex items-center gap-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">{{ $t('settings.mode') }}</p>
            <p class="text-xs text-muted">{{ $t('settings.modeDescription') }}</p>
          </div>
          <USelectMenu
            v-model="colorMode.preference"
            :items="modes"
            value-key="value"
            :search-input="false"
            :aria-label="$t('settings.mode')"
            class="w-56"
          />
        </li>
      </ul>
    </section>

    <section class="mb-8">
      <div class="flex items-center gap-4">
        <div class="min-w-0 flex-1">
          <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.snippets') }}</h2>
          <p class="text-xs text-muted">{{ $t('settings.snippetsDescription') }}</p>
        </div>
        <UButton
          :label="$t('settings.addSnippet')"
          :icon="icons.noteAdd"
          color="neutral"
          variant="outline"
          :disabled="!me"
          @click="addSnippet"
        />
      </div>
      <ul>
        <li
          v-for="(snippet, index) in me?.snippets"
          :key="index"
          class="flex items-center gap-3 border-b border-muted py-3 last:border-b-0"
        >
          <UIcon :name="snippet.template ? icons.template : icons.snippet" class="size-4 shrink-0 text-muted" />
          <p class="min-w-0 flex-1 truncate text-sm">{{ snippet.name }}</p>
          <UBadge
            :label="snippet.template ? $t('settings.template') : $t('settings.snippet')"
            color="neutral"
            variant="subtle"
          />
          <UButton
            :icon="icons.rename"
            color="neutral"
            variant="ghost"
            :aria-label="$t('settings.editSnippet')"
            @click="changeSnippet(index)"
          />
          <UButton
            :icon="icons.delete"
            color="neutral"
            variant="ghost"
            :aria-label="$t('common.delete')"
            @click="removeSnippet(index)"
          />
        </li>
        <li v-if="me && !me.snippets.length" class="py-3 text-xs text-muted">{{ $t('settings.snippetsEmpty') }}</li>
      </ul>
    </section>

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.trash') }}</h2>
      <div class="flex items-center gap-4 py-3">
        <div class="min-w-0 flex-1">
          <p class="text-sm">{{ $t('settings.retention') }}</p>
          <p class="text-xs text-muted">{{ $t('settings.retentionDescription') }}</p>
        </div>
        <USelectMenu
          :model-value="me?.trash_retention_days"
          :items="retentions"
          value-key="value"
          :search-input="false"
          :aria-label="$t('settings.retention')"
          class="w-56"
          @update:model-value="changeRetention"
        />
      </div>
    </section>

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.data') }}</h2>
      <ul>
        <li class="flex items-center gap-4 border-b border-muted py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">{{ $t('settings.export') }}</p>
            <p class="text-xs text-muted">{{ $t('settings.exportDescription') }}</p>
          </div>
          <UButton
            :label="$t('settings.export')"
            :icon="icons.export"
            color="neutral"
            variant="outline"
            :href="exportURL"
            download
          />
        </li>
        <li class="flex items-center gap-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">{{ $t('settings.import') }}</p>
            <p class="text-xs text-muted">{{ $t('settings.importDescription') }}</p>
          </div>
          <USelectMenu
            v-model="parent"
            :items="destinations"
            value-key="value"
            :aria-label="$t('settings.importInto')"
            class="w-56"
          />
          <input ref="input" type="file" accept=".zip,application/zip" class="hidden" @change="importZip" />
          <UButton
            :label="$t('settings.import')"
            :icon="icons.import"
            color="neutral"
            variant="outline"
            :loading="importing"
            @click="input?.click()"
          />
        </li>
      </ul>
    </section>

    <section>
      <h2 class="mb-1 text-sm font-semibold">{{ $t('settings.danger') }}</h2>
      <div class="flex items-center gap-4 rounded-md border border-error/25 bg-error/5 px-4 py-3">
        <div class="min-w-0 flex-1">
          <p class="text-sm">{{ $t('settings.deleteData') }}</p>
          <p class="text-xs text-muted">{{ $t('settings.deleteDataDescription') }}</p>
        </div>
        <UButton :label="$t('common.delete')" color="error" variant="ghost" :disabled="!me" @click="deleteData" />
      </div>
    </section>
  </div>
</template>
