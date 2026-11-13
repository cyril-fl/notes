<script setup lang="ts">
import type { SavedSearch } from '~/types/items';

const props = defineProps<{ title: string; value?: SavedSearch }>();
const emit = defineEmits<{ close: [value?: SavedSearch] }>();

const { t } = useI18n();
const { data: tags } = useTags();
const destinations = useDestinations();
const folderOptions = computed(() => destinations.value.filter((d) => d.value !== ROOT));
const tagOptions = computed(() => tags.value.map((tag) => ({ label: `#${tag.value}`, value: tag.id })));
const tagModeOptions = [
  { label: t('search.saved.tagModeAny'), value: 'any' },
  { label: t('search.saved.tagModeAll'), value: 'all' },
];
const folderModeOptions = [
  { label: t('search.saved.folderModeInclude'), value: 'include' },
  { label: t('search.saved.folderModeExclude'), value: 'exclude' },
];

// Date bounds are shown as a plain day (<input type="date">), sent as a day boundary in UTC.
function toDay(iso?: string) {
  return iso ? iso.slice(0, 10) : '';
}
function fromDay(day: string, end: boolean) {
  return day ? `${day}T${end ? '23:59:59.999Z' : '00:00:00.000Z'}` : undefined;
}

const name = ref(props.value?.name ?? '');
const tagIds = ref(props.value?.tag_ids ?? []);
const tagMode = ref(props.value?.tag_mode ?? 'any');
const folderIds = ref(props.value?.folder_ids ?? []);
const folderMode = ref(props.value?.folder_mode ?? 'include');
const createdFrom = ref(toDay(props.value?.created_from));
const createdTo = ref(toDay(props.value?.created_to));
const editedFrom = ref(toDay(props.value?.edited_from));
const editedTo = ref(toDay(props.value?.edited_to));
const text = ref(props.value?.text ?? '');

function submit() {
  if (!name.value.trim()) return;
  emit('close', {
    id: props.value?.id ?? '',
    name: name.value.trim(),
    tag_ids: tagIds.value,
    tag_mode: tagMode.value,
    folder_ids: folderIds.value,
    folder_mode: folderMode.value,
    created_from: fromDay(createdFrom.value, false),
    created_to: fromDay(createdTo.value, true),
    edited_from: fromDay(editedFrom.value, false),
    edited_to: fromDay(editedTo.value, true),
    text: text.value.trim(),
  });
}
</script>

<template>
  <UModal :title="title" :close="false" :ui="{ content: 'max-w-lg' }" @update:open="emit('close')">
    <template #body>
      <form id="saved-search" class="flex flex-col gap-4" @submit.prevent="submit">
        <UInput
          v-model="name"
          autofocus
          class="w-full"
          :placeholder="$t('search.saved.name')"
          :aria-label="$t('search.saved.name')"
        />

        <div class="flex flex-col gap-2">
          <label class="text-sm text-muted">{{ $t('common.tags') }}</label>
          <div class="flex gap-2">
            <USelectMenu
              v-model="tagIds"
              multiple
              :items="tagOptions"
              value-key="value"
              class="flex-1"
              :placeholder="$t('search.saved.anyTag')"
              :aria-label="$t('common.tags')"
            />
            <USelect
              v-if="tagIds.length > 1"
              v-model="tagMode"
              :items="tagModeOptions"
              value-key="value"
              class="w-40"
              :aria-label="$t('search.saved.tagMode')"
            />
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <label class="text-sm text-muted">{{ $t('search.saved.folders') }}</label>
          <div class="flex gap-2">
            <USelectMenu
              v-model="folderIds"
              multiple
              :items="folderOptions"
              value-key="value"
              class="flex-1"
              :placeholder="$t('search.saved.anyFolder')"
              :aria-label="$t('search.saved.folders')"
            />
            <USelect
              v-if="folderIds.length"
              v-model="folderMode"
              :items="folderModeOptions"
              value-key="value"
              class="w-40"
              :aria-label="$t('search.saved.folderMode')"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div class="flex flex-col gap-2">
            <label class="text-sm text-muted">{{ $t('search.saved.created') }}</label>
            <div class="flex gap-2">
              <UInput v-model="createdFrom" type="date" class="w-full" :aria-label="$t('search.saved.createdFrom')" />
              <UInput v-model="createdTo" type="date" class="w-full" :aria-label="$t('search.saved.createdTo')" />
            </div>
          </div>
          <div class="flex flex-col gap-2">
            <label class="text-sm text-muted">{{ $t('search.saved.edited') }}</label>
            <div class="flex gap-2">
              <UInput v-model="editedFrom" type="date" class="w-full" :aria-label="$t('search.saved.editedFrom')" />
              <UInput v-model="editedTo" type="date" class="w-full" :aria-label="$t('search.saved.editedTo')" />
            </div>
          </div>
        </div>

        <UInput
          v-model="text"
          class="w-full"
          :placeholder="$t('search.saved.textPlaceholder')"
          :aria-label="$t('search.saved.text')"
        />
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton type="submit" form="saved-search" :label="$t('settings.save')" :disabled="!name.trim()" />
      </div>
    </template>
  </UModal>
</template>
