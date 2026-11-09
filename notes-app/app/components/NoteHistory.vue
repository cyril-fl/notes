<script setup lang="ts">
import type { Item, ItemDetail, NoteVersion } from '~/types/items';

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
// The API keeps a version at the first save after a 15 minute pause, at least one an hour, and on a restore:
// read the list again only then, not on each save (#97, #99).
let restored = false;
watch(
  () => props.savedAt,
  (savedAt, previous) => {
    const newest = versions.value[0]?.created_at;
    if (!props.panel || !savedAt) return;
    const since = (date?: string) => (date ? Date.parse(savedAt) - Date.parse(date) : Infinity);
    if (restored || since(previous) >= 15 * 60_000 || since(newest) >= 60 * 60_000) {
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
  text.value !== undefined && current.value !== undefined ? lineChanges(text.value, current.value) : undefined,
);

// The list has no texts (#99): the selected version's is read, the answer of an older click dropped.
const text = ref<string>();
watch(selected, async (version) => {
  text.value = undefined;
  if (!version) return;
  const read = await api<NoteVersion>(`/api/notes/${props.noteId}/versions/${version.id}`);
  if (selected.value === version) text.value = read.content ?? '';
});

function restore() {
  if (text.value === undefined) return;
  restored = true;
  emit('restore', text.value);
  open.value = false;
}

// A fork: the selected version as a new note beside this one, with its tags, titled after its date so its links stay apart.
async function fork() {
  if (text.value === undefined || !selected.value) return;
  const note = await api<ItemDetail>(`/api/notes/${props.noteId}`);
  const date = new Date(selected.value.created_at).toLocaleString(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
  const title = t('history.forkTitle', { title: displayTitle(note.item.title, t('untitled')), date });
  const item = await api<Item>('/api/notes', {
    method: 'POST',
    body: { content: retitle(text.value, title), parent_id: note.item.parent_id, tags: note.tags },
    invalid: { [NO_PARENT]: t('items.parentInTrash') },
  });
  open.value = false;
  await navigateTo(`/notes/${item.id}`);
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
          >{{ text }}</pre>
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
        <UButton
          :label="$t('history.fork')"
          color="neutral"
          variant="outline"
          :disabled="text === undefined"
          @click="fork"
        />
        <UButton :label="$t('history.restore')" :disabled="text === undefined || readonly" @click="restore" />
      </div>
    </template>
  </UModal>
</template>
