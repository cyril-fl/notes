<script lang="ts" setup>
interface Props {
  href?: string;
}

interface Emits {
  (e: 'close', result: { href: string } | { unset: true } | undefined): void;
}

const props = withDefaults(defineProps<Props>(), {
  href: '',
});

const emit = defineEmits<Emits>();

const { t } = useI18n();
const { icons } = useIcons();
const isEditing = computed(() => !!props.href);

const url = ref(props.href ?? '');

const ALLOWED_PROTOCOLS = ['http', 'https', 'mailto'];

const isValidUrl = computed(() => {
  const trimmed = url.value.trim();
  if (!trimmed) return false;
  const protocol = (trimmed.split(':')[0] ?? '').toLowerCase();
  return ALLOWED_PROTOCOLS.includes(protocol) || !trimmed.includes(':');
});

const handleSubmit = () => {
  if (!isValidUrl.value) return;
  emit('close', { href: url.value.trim() });
};

const handleRemove = () => {
  emit('close', { unset: true });
};

const handleCancel = () => {
  emit('close', undefined);
};
</script>

<template>
  <UModal
    :open="true"
    @update:open="
      (val: boolean) => {
        if (!val) handleCancel();
      }
    "
  >
    <template #content>
      <div class="flex justify-between items-center gap-2 p-4">
        <UInput
          v-model="url"
          autofocus
          :placeholder="t('editor.link.placeholder')"
          :icon="icons.link"
          class="grow"
          @keydown.enter="handleSubmit"
        />

        <div class="flex gap-2">
          <UButton
            v-if="isEditing"
            color="error"
            variant="soft"
            :icon="icons.unlink"
            @click="handleRemove"
          >
            {{ t('editor.link.remove') }}
          </UButton>

          <UButton color="neutral" variant="ghost" @click="handleCancel">
            {{ t('editor.link.cancel') }}
          </UButton>

          <UButton
            color="primary"
            :disabled="!isValidUrl"
            @click="handleSubmit"
          >
            {{ t('editor.link.submit') }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
