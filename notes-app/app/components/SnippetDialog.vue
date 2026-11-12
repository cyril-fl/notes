<script setup lang="ts">
import type { Snippet } from '~/types/items';

const props = defineProps<{ title: string; value?: Snippet }>();
const emit = defineEmits<{ close: [value?: Snippet] }>();

const name = ref(props.value?.name ?? '');
const content = ref(props.value?.content ?? '');
// The editor starts from it and keeps its own text: handing back what it writes would move the cursor.
const start = content.value;

function submit() {
  if (name.value.trim()) emit('close', { name: name.value.trim(), content: content.value });
}
</script>

<template>
  <UModal :title="title" :close="false" :ui="{ content: 'max-w-2xl' }" @update:open="emit('close')">
    <template #body>
      <form id="snippet" class="flex flex-col gap-4" @submit.prevent="submit">
        <UInput
          v-model="name"
          autofocus
          class="w-full"
          :placeholder="$t('settings.snippetName')"
          :aria-label="$t('settings.snippetName')"
        />
        <UFormField :help="$t('settings.snippetHelp')">
          <NoteWriter
            :model-value="start"
            compact
            :label="$t('settings.snippetContent')"
            class="h-80 rounded-md ring ring-accented"
            @update:model-value="content = $event"
          />
        </UFormField>
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton type="submit" form="snippet" :label="$t('settings.save')" :disabled="!name.trim()" />
      </div>
    </template>
  </UModal>
</template>
