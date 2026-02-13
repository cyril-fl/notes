<script lang="ts" setup>
/* Define */
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

const isEditing = computed(() => !!props.href);

/* Data */
const url = ref(props.href ?? '');

/* Methods */
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
    :title="isEditing ? 'Modifier le lien' : 'Ajouter un lien'"
    @update:open="
      (val: boolean) => {
        if (!val) handleCancel();
      }
    "
  >
    <template #body>
      <form @submit.prevent="handleSubmit">
        <UInput
          v-model="url"
          autofocus
          placeholder="https://exemple.com"
          icon="i-lucide-link"
          size="lg"
        />
      </form>
    </template>

    <template #footer>
      <div class="flex items-center gap-2 w-full">
        <UButton
          v-if="isEditing"
          color="error"
          variant="soft"
          icon="i-lucide-unlink"
          @click="handleRemove"
        >
          Supprimer
        </UButton>

        <div class="grow" />

        <UButton color="neutral" variant="ghost" @click="handleCancel">
          Annuler
        </UButton>

        <UButton color="primary" :disabled="!isValidUrl" @click="handleSubmit">
          Valider
        </UButton>
      </div>
    </template>
  </UModal>
</template>
