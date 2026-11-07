<script setup lang="ts">
defineProps<{ title: string; description?: string; action: string }>();
const emit = defineEmits<{ close: [confirmed?: boolean] }>();
</script>

<template>
  <UModal :title="title" :description="description" :ui="{ content: 'max-w-sm' }" :close="false" @update:open="emit('close')">
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton :label="$t('common.cancel')" color="neutral" variant="ghost" @click="emit('close')" />
        <UButton :label="action" color="error" @click="emit('close', true)" />
      </div>
    </template>
  </UModal>
</template>
