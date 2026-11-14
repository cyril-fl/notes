<script setup lang="ts">
import type { Neighbors } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

// panel: in the context panel, open, without the graph link: the mini graph below has its own, and shown even empty once read, like the history below.
const props = defineProps<{ noteId: string; panel?: boolean }>();

const { has } = usePro();
const { data: neighbors, status } = useNeighbors(() => props.noteId);
// The related notes of the previous note until this one's come, rather than an empty section (#154).
const last = useState<Neighbors | null | undefined>('related-shown');
watch(neighbors, (value) => value && (last.value = value), { immediate: true });
const data = computed(() => neighbors.value ?? (status.value === 'error' ? undefined : last.value));

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
  <UCollapsible v-if="sections.length || (panel && data)" :default-open="panel">
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
        <NuxtLink
          v-if="has('graph') && !panel"
          :to="{ path: '/graph', query: { note: noteId } }"
          class="ml-auto text-xs text-primary"
          @click.stop
          >{{ $t('related.showGraph') }}</NuxtLink
        >
      </div>
    </template>
    <template #content>
      <UEmpty
        v-if="!sections.length"
        :icon="icons.link"
        :title="$t('related.emptyTitle')"
        :description="$t('related.emptyDescription')"
        size="xs"
        variant="naked"
        :ui="{ root: 'min-h-0 gap-0 p-2 sm:p-2 lg:p-2', header: 'gap-1', avatar: 'mb-0 size-6 text-sm' }"
        class="mb-4"
      />
      <div v-else class="max-h-[40vh] space-y-6 overflow-y-auto pb-4">
        <div v-for="section in sections" :key="section.title">
          <h2 class="mb-2 text-xs font-medium text-muted">{{ section.title }}</h2>
          <NoteList :notes="section.notes" />
        </div>
      </div>
    </template>
  </UCollapsible>
</template>
