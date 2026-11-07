<script setup lang="ts">
import type { Snippet } from '~/types/items';

const props = defineProps<{ title: string; value?: Snippet }>();
const emit = defineEmits<{ close: [value?: Snippet] }>();

const { t } = useI18n();

const snippet = ref<Snippet>({ name: '', content: '', template: false, ...props.value });
const kinds = computed(() => [
  { label: t('settings.snippet'), value: false },
  { label: t('settings.template'), value: true },
]);

function submit() {
  const name = snippet.value.name.trim();
  if (name) emit('close', { ...snippet.value, name });
}
</script>

<template>
  <UModal :title="title" :close="false" @update:open="emit('close')">
    <template #body>
      <form id="snippet" class="flex flex-col gap-4" @submit.prevent="submit">
        <div class="flex gap-2">
          <UInput
            v-model="snippet.name"
            autofocus
            class="flex-1"
            :placeholder="$t('settings.snippetName')"
            :aria-label="$t('settings.snippetName')"
          />
          <USelectMenu
            v-model="snippet.template"
            :items="kinds"
            value-key="value"
            :search-input="false"
            :aria-label="$t('settings.snippetKind')"
            class="w-36"
          />
        </div>
        <UFormField :help="$t('settings.snippetHelp')">
          <UTextarea
            v-model="snippet.content"
            :rows="8"
            autoresize
            class="w-full font-mono"
            :aria-label="$t('settings.snippetContent')"
          />
        </UFormField>
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton type="submit" form="snippet" :label="$t('settings.save')" :disabled="!snippet.name.trim()" />
      </div>
    </template>
  </UModal>
</template>
