<script setup lang="ts">
import type { NoteVersion } from '~/types/items';

// Pro (#61): the texts the API kept before the edits of the note, newest first. Restoring hands one to the editor, which saves it as any edit.
const props = defineProps<{ noteId: string; readonly?: boolean }>();
const emit = defineEmits<{ restore: [content: string] }>();

const { icons } = useIcons();
const { locale } = useI18n();
const api = useApi();

const open = ref(false);
const versions = ref<NoteVersion[]>([]);
const selected = ref<NoteVersion>();

watch(open, async (opened) => {
  if (!opened) return;
  versions.value = await api<NoteVersion[]>(`/api/notes/${props.noteId}/versions`);
  selected.value = versions.value[0];
});

const date = (version: NoteVersion) =>
  new Date(version.created_at).toLocaleString(locale.value, { dateStyle: 'medium', timeStyle: 'short' });

function restore() {
  if (!selected.value) return;
  emit('restore', selected.value.content);
  open.value = false;
}
</script>

<template>
  <UModal v-model:open="open" :title="$t('history.title')" :ui="{ content: 'max-w-3xl' }">
    <UTooltip :text="$t('history.title')">
      <UButton :icon="icons.history" color="neutral" variant="ghost" size="xs" :aria-label="$t('history.title')" />
    </UTooltip>
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
