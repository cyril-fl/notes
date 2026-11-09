<script setup lang="ts">
import type { ItemDetail, NoteVersion } from '~/types/items';

// Pro (#61): the texts the API kept before the edits of the note, newest first. Restoring hands one to the editor, which saves it as any edit.
// panel: the last versions listed in the context panel; otherwise a button under the text (no panel on a phone).
// savedAt: the updated_at of the note, the list reads again as it changes.
const props = defineProps<{ noteId: string; readonly?: boolean; panel?: boolean; savedAt?: string }>();
const emit = defineEmits<{ restore: [content: string] }>();

const { icons } = useIcons();
const { locale, t } = useI18n();
const api = useApi();

const open = ref(false);
const versions = ref<NoteVersion[]>([]);
const selected = ref<NoteVersion>();
// The panel shows the last few, the dialog all of them.
const shown = 5;

async function load() {
  versions.value = await api<NoteVersion[]>(`/api/notes/${props.noteId}/versions`, { quiet: true });
}
if (props.panel) load();
// The API keeps a version at most every 5 minutes, or on a restore: read the full texts again only then, not on each save (#97).
let restored = false;
watch(
  () => props.savedAt,
  (savedAt) => {
    const newest = versions.value[0]?.created_at;
    if (!props.panel || !savedAt) return;
    if (restored || !newest || Date.parse(savedAt) - Date.parse(newest) >= 5 * 60_000) {
      restored = false;
      load();
    }
  },
);

async function show(version?: NoteVersion) {
  open.value = true;
  current.value = undefined;
  await load();
  selected.value = versions.value.find((other) => other.id === version?.id) ?? versions.value[0];
}

// A text refused by a 409 says so (#78).
const label = (version: NoteVersion) =>
  new Date(version.created_at).toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' }) +
  (version.conflict ? ` · ${t('history.conflict')}` : '');

// The comparison (#78): the version against the text as saved, read when asked; lines of the version only struck, of the text only underlined.
const current = ref<string>();
async function compare() {
  if (current.value !== undefined) return void (current.value = undefined);
  current.value = (await api<ItemDetail>(`/api/notes/${props.noteId}`)).content ?? '';
}
const changes = computed(() =>
  selected.value && current.value !== undefined ? lineChanges(selected.value.content, current.value) : undefined,
);

function restore() {
  if (!selected.value) return;
  restored = true;
  emit('restore', selected.value.content);
  open.value = false;
}
</script>

<template>
  <!-- Folds like the related notes above it. -->
  <UCollapsible v-if="panel" default-open as="section" :aria-label="$t('history.title')">
    <template #default="{ open: unfolded }">
      <div class="flex items-center py-2">
        <UButton
          :label="$t('history.panelTitle', { count: versions.length })"
          :trailing-icon="unfolded ? icons.chevronDown : icons.chevronUp"
          color="neutral"
          variant="ghost"
          size="xs"
          class="-ml-2"
        />
      </div>
    </template>
    <template #content>
      <div class="pb-4">
        <p v-if="!versions.length" class="text-xs text-dimmed">{{ $t('history.empty') }}</p>
        <button
          v-for="version in versions.slice(0, shown)"
          :key="version.id"
          type="button"
          class="block w-full truncate rounded py-1 text-left text-sm text-muted hover:text-default"
          @click="show(version)"
        >
          {{ label(version) }}
        </button>
        <button v-if="versions.length > shown" type="button" class="py-1 text-xs text-primary" @click="show()">
          {{ $t('history.all', { count: versions.length }) }}
        </button>
      </div>
    </template>
  </UCollapsible>
  <UTooltip v-else :text="$t('history.title')">
    <UButton
      :icon="icons.history"
      color="neutral"
      variant="ghost"
      size="xs"
      :aria-label="$t('history.title')"
      @click="show()"
    />
  </UTooltip>

  <UModal v-model:open="open" :title="$t('history.title')" :ui="{ content: 'max-w-3xl' }">
    <template #body>
      <p v-if="!versions.length" class="text-sm text-muted">{{ $t('history.empty') }}</p>
      <div v-else class="flex flex-col gap-4 sm:flex-row">
        <nav class="flex shrink-0 flex-col sm:w-48" :aria-label="$t('history.title')">
          <UButton
            v-for="version in versions"
            :key="version.id"
            :label="label(version)"
            color="neutral"
            :variant="version === selected ? 'soft' : 'ghost'"
            size="sm"
            :aria-pressed="version === selected"
            @click="selected = version"
          />
        </nav>
        <div v-if="changes" class="max-h-[60vh] min-w-0 flex-1 overflow-y-auto text-sm">
          <component
            :is="{ same: 'span', removed: 'del', added: 'ins' }[change.kind]"
            v-for="(change, index) in changes"
            :key="index"
            class="block whitespace-pre-wrap"
            :class="{ same: '', removed: 'bg-error/10 text-error', added: 'bg-success/10 text-success' }[change.kind]"
            >{{ change.lines.join('\n') }}</component
          >
        </div>
        <pre
          v-else-if="selected"
          class="max-h-[60vh] min-w-0 flex-1 overflow-y-auto font-sans text-sm whitespace-pre-wrap"
          >{{ selected.content }}</pre>
      </div>
    </template>
    <template #footer>
      <div class="flex w-full flex-wrap justify-end gap-2">
        <UButton
          :label="$t('history.compare')"
          color="neutral"
          variant="outline"
          class="mr-auto"
          :disabled="!selected"
          :aria-pressed="current !== undefined"
          @click="compare"
        />
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="open = false" />
        <UButton :label="$t('history.restore')" :disabled="!selected || readonly" @click="restore" />
      </div>
    </template>
  </UModal>
</template>
