<script setup lang="ts">
import type { Snippet, User } from '~/types/items';

// Decided before the page renders, so the Pro mode never shows every section first and then jumps to one.
definePageMeta({
  middleware: async (to) => {
    // Not useApi: its useI18n needs a component. Without a session the page goes on, and the layout sends to /login.
    // Kept as the layout's me, so the sidebar draws the settings nav from the start too.
    const cached = useNuxtData<User>('me').data;
    cached.value ??= await $fetch<User>('/api/me', {
      baseURL: useRuntimeConfig().public.apiBase,
      credentials: 'include',
    }).catch(() => undefined);
    const me = cached.value;
    // A page per section is the Pro mode's: without it, /settings shows them all; with it, /settings opens the first.
    const section = to.params.section as string | undefined;
    if (me && unlocked(features.settingsNav, planOf(me))) {
      if (!settingsSections.some(({ id }) => id === section)) return navigateTo('/settings/general', { replace: true });
    } else if (section) return navigateTo('/settings', { replace: true });
  },
});

const { icons } = useIcons();
const { t, locale } = useI18n();

// /settings shows every section; /settings/<section> one, the page of the Pro mode's settings nav.
const route = useRoute();
const current = computed(() => route.params.section as SettingsSection | undefined);
const show = (id: SettingsSection) => !current.value || current.value === id;

const api = useApi();
const toast = useToast();
// The API sends the zip as an attachment: a plain link downloads it.
const exportURL = `${useRuntimeConfig().public.apiBase}/api/export`;

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

const { pro } = usePro();

// The settings switch between one page and a page per section: on, the switch stays in view on its own page.
async function changePro(next: boolean) {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { pro: next } });
  await navigateTo(next ? '/settings/pro' : '/settings', { replace: true });
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

const { confirm, snippet: editSnippet, importNotes } = useDialogs();

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
</script>

<template>
  <div class="p-8">
    <PageHeader
      :title="current ? $t(`settings.sections.${current}`) : $t('common.settings')"
      :subtitle="current ? $t('common.settings') : undefined"
    />

    <div class="space-y-10">
      <section v-if="show('general')">
        <SettingsHeading id="general" />
        <ul class="divide-y divide-default rounded-lg border border-default">
          <li class="flex items-center gap-4 px-4 py-3">
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
          </li>
          <li class="flex items-center gap-4 px-4 py-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm">{{ $t('settings.language') }}</p>
              <p class="text-xs text-muted">{{ $t('settings.languageDescription') }}</p>
            </div>
            <USelectMenu
              :model-value="locale"
              :items="languages"
              value-key="value"
              :search-input="false"
              :aria-label="$t('settings.language')"
              class="w-56"
              @update:model-value="changeLanguage"
            />
          </li>
        </ul>
      </section>

      <section v-if="show('appearance')">
        <SettingsHeading id="appearance" />
        <ul class="divide-y divide-default rounded-lg border border-default">
          <li class="px-4 py-3">
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
          <li class="flex items-center gap-4 px-4 py-3">
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

      <section v-if="show('editor')">
        <SettingsHeading id="editor" />
        <ul class="divide-y divide-default rounded-lg border border-default">
          <li class="flex items-center gap-4 px-4 py-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm">{{ $t('settings.snippets') }}</p>
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
          </li>
          <li v-for="(snippet, index) in me?.snippets" :key="index" class="flex items-center gap-3 px-4 py-2">
            <UIcon :name="icons.snippet" class="size-4 shrink-0 text-muted" />
            <p class="min-w-0 flex-1 truncate text-sm">{{ snippet.name }}</p>
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
          <li v-if="me && !me.snippets.length" class="px-4 py-3 text-xs text-muted">
            {{ $t('settings.snippetsEmpty') }}
          </li>
        </ul>
      </section>

      <section v-if="show('data')">
        <SettingsHeading id="data" />
        <ul class="divide-y divide-default rounded-lg border border-default">
          <li class="flex items-center gap-4 px-4 py-3">
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
          <li class="flex items-center gap-4 px-4 py-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm">{{ $t('settings.import') }}</p>
              <p class="text-xs text-muted">{{ $t('settings.importDescription') }}</p>
            </div>
            <UButton
              :label="$t('settings.import')"
              :icon="icons.import"
              color="neutral"
              variant="outline"
              @click="importNotes()"
            />
          </li>
          <li class="flex items-center gap-4 px-4 py-3">
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
          </li>
        </ul>
        <!-- Each section keeps what cannot be undone at its bottom, apart. -->
        <h3 class="mt-6 mb-2 text-xs font-medium text-error">{{ $t('settings.danger') }}</h3>
        <div class="flex items-center gap-4 rounded-lg border border-error/25 bg-error/5 px-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm">{{ $t('settings.deleteData') }}</p>
            <p class="text-xs text-muted">{{ $t('settings.deleteDataDescription') }}</p>
          </div>
          <UButton :label="$t('common.delete')" color="error" variant="ghost" :disabled="!me" @click="deleteData" />
        </div>
      </section>

      <section v-if="show('pro')">
        <SettingsHeading id="pro" />
        <div class="overflow-hidden rounded-lg border border-default">
          <div class="flex items-center gap-4 border-b border-default px-4 py-3">
            <p class="min-w-0 flex-1 text-sm">{{ $t('settings.proMode') }}</p>
            <USwitch
              :model-value="pro"
              :disabled="!me"
              :aria-label="$t('settings.proMode')"
              @update:model-value="changePro"
            />
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-default">
                <th class="px-4 py-2 text-left text-xs font-normal text-muted">{{ $t('settings.feature') }}</th>
                <th class="w-24 px-2 py-2 font-medium">
                  <div class="flex flex-col items-center gap-1">
                    {{ $t('settings.standard') }}
                    <UBadge v-if="me && !pro" :label="$t('settings.currentPlan')" size="sm" variant="subtle" />
                  </div>
                </th>
                <th class="w-24 bg-primary/5 px-2 py-2 font-medium text-primary">
                  <div class="flex flex-col items-center gap-1">
                    <span class="inline-flex items-center gap-1"
                      ><UIcon :name="icons.pro" />{{ $t('settings.advanced') }}</span
                    >
                    <UBadge v-if="pro" :label="$t('settings.currentPlan')" size="sm" variant="subtle" />
                  </div>
                </th>
              </tr>
            </thead>
            <tbody v-for="{ group, features: listed } in listedFeatures" :key="group">
              <tr class="border-b border-default bg-elevated/50">
                <th colspan="3" class="px-4 py-1.5 text-left text-xs font-medium text-muted">
                  {{ $t(`settings.featureGroups.${group}`) }}
                </th>
              </tr>
              <tr v-for="feature in listed" :key="feature" class="border-b border-default last:border-b-0">
                <td class="px-4 py-2 text-xs">{{ $t(`features.${feature}`) }}</td>
                <td class="px-2 text-center">
                  <UIcon
                    v-if="features[feature] === 'base'"
                    :name="icons.current"
                    :aria-label="$t('settings.included')"
                    class="size-4 text-primary"
                  />
                  <UIcon
                    v-else
                    :name="icons.divider"
                    :aria-label="$t('settings.notIncluded')"
                    class="size-4 text-dimmed"
                  />
                </td>
                <td class="bg-primary/5 px-2 text-center">
                  <UIcon :name="icons.current" :aria-label="$t('settings.included')" class="size-4 text-primary" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </div>
</template>
