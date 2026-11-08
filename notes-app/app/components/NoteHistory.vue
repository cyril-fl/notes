<script setup lang="ts">
import type { NoteVersion } from '~/types/items';

// Pro (#61): the texts the API kept before the edits of the note, newest first. Restoring hands one to the editor, which saves it as any edit.
// panel: the last versions listed in the context panel; otherwise a button under the text (no panel on a phone).
// savedAt: the updated_at of the note, the list reads again as it changes.
const props = defineProps<{ noteId: string; readonly?: boolean; panel?: boolean; savedAt?: string }>();
const emit = defineEmits<{ restore: [content: string] }>();

const { icons } = useIcons();
const { locale } = useI18n();
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
// A version is kept at most every 5 minutes: no need to read again on each save.
watchThrottled(
  () => props.savedAt,
  () => props.panel && load(),
  { throttle: 30_000 },
);

async function show(version?: NoteVersion) {
  open.value = true;
  await load();
  selected.value = versions.value.find((other) => other.id === version?.id) ?? versions.value[0];
}

const date = (version: NoteVersion) =>
  new Date(version.created_at).toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' });

function restore() {
  if (!selected.value) return;
  emit('restore', selected.value.content);
  open.value = false;
}
</script>

<template>
  <section v-if="panel" :aria-label="$t('history.title')">
    <h2 class="mb-2 text-xs font-medium text-muted">{{ $t('history.title') }}</h2>
    <p v-if="!versions.length" class="text-xs text-dimmed">{{ $t('history.empty') }}</p>
    <button
      v-for="version in versions.slice(0, shown)"
      :key="version.id"
      type="button"
      class="block w-full truncate rounded py-1 text-left text-sm text-muted hover:text-default"
      @click="show(version)"
    >
      {{ date(version) }}
    </button>
    <button v-if="versions.length > shown" type="button" class="py-1 text-xs text-primary" @click="show()">
      {{ $t('history.all', { count: versions.length }) }}
    </button>
  </section>
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
            :label="date(version)"
            color="neutral"
            :variant="version === selected ? 'soft' : 'ghost'"
            size="sm"
            :aria-pressed="version === selected"
            @click="selected = version"
          />
        </nav>
        <pre
          v-if="selected"
          class="max-h-[60vh] min-w-0 flex-1 overflow-y-auto font-sans text-sm whitespace-pre-wrap"
          >{{ selected.content }}</pre>
      </div>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="open = false" />
        <UButton :label="$t('history.restore')" :disabled="!selected || readonly" @click="restore" />
      </div>
    </template>
  </UModal>
</template>
