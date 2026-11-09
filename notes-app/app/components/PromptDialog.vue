<script setup lang="ts">
const props = defineProps<{
  title: string;
  value?: string;
  placeholder?: string;
  action?: string;
  multiline?: boolean;
}>();
const emit = defineEmits<{ close: [value?: string] }>();

const text = ref(props.value ?? '');

function submit() {
  if (text.value.trim()) emit('close', text.value.trim());
}
</script>

<template>
  <UModal :title="title" :ui="{ content: 'max-w-sm' }" :close="false" @update:open="emit('close')">
    <template #body>
      <form id="prompt" @submit.prevent="submit">
        <!-- Several lines: Enter goes to the next one, Ctrl or ⌘ Enter confirms. -->
        <UTextarea
          v-if="multiline"
          v-model="text"
          autofocus
          autoresize
          :rows="4"
          class="w-full font-mono"
          :placeholder="placeholder"
          :aria-label="title"
          @keydown.enter.ctrl.prevent="submit"
          @keydown.enter.meta.prevent="submit"
        />
        <UInput v-else v-model="text" autofocus class="w-full" :placeholder="placeholder" :aria-label="title" />
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton type="submit" form="prompt" :label="action ?? $t('common.confirm')" :disabled="!text.trim()" />
      </div>
    </template>
  </UModal>
</template>
