<script lang="ts" setup>
interface Props {
  open: boolean;
  itemType: 'note' | 'folder';
}

interface Emits {
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();
const { t } = useI18n();
</script>

<template>
  <UModal
    :open="open"
    @update:open="
      (val: boolean) => {
        if (!val) emit('cancel');
      }
    "
  >
    <template #content>
      <div class="p-6 space-y-4">
        <p class="text-sm text-default">
          {{
            itemType === 'folder'
              ? t('modal.confirm_delete.folder')
              : t('modal.confirm_delete.note')
          }}
        </p>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="ghost" @click="emit('cancel')">
            {{ t('modal.confirm_delete.cancel') }}
          </UButton>
          <UButton color="error" @click="emit('confirm')">
            {{ t('modal.confirm_delete.confirm') }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
