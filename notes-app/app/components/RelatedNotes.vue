<script setup lang="ts">
import type { Neighbors } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

const props = defineProps<{ noteId: string }>();

const api = useApi();
const { data } = useLazyAsyncData(`neighbors:${props.noteId}`, () =>
  api<Neighbors>(`/api/notes/${props.noteId}/neighbors`, { quiet: true }),
);

const sections = computed(() =>
  [
    { title: t('related.links'), notes: data.value?.links ?? [] },
    { title: t('related.backlinks'), notes: data.value?.backlinks ?? [] },
    { title: t('related.sameTags'), notes: data.value?.neighbors ?? [] },
  ].filter((section) => section.notes.length),
);

const total = computed(() => sections.value.reduce((sum, section) => sum + section.notes.length, 0));
</script>

<template>
  <UCollapsible v-if="sections.length">
    <template #default="{ open }">
      <div class="flex items-center py-2">
        <UButton
          :label="$t('related.title', { total })"
          :trailing-icon="open ? icons.chevronDown : icons.chevronUp"
          color="neutral"
          variant="ghost"
          size="xs"
          class="-ml-2"
        />
        <NuxtLink :to="{ path: '/graph', query: { note: noteId } }" class="ml-auto text-xs text-primary" @click.stop>{{
          $t('related.showGraph')
        }}</NuxtLink>
      </div>
    </template>
    <template #content>
      <div class="max-h-[40vh] space-y-6 overflow-y-auto pb-4">
        <div v-for="section in sections" :key="section.title">
          <h2 class="mb-2 text-xs font-medium text-muted">{{ section.title }}</h2>
          <NoteList :notes="section.notes" />
        </div>
      </div>
    </template>
  </UCollapsible>
</template>
