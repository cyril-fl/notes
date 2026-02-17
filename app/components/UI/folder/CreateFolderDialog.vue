<script lang="ts" setup>
const { $hooks } = useNuxtApp();
const { t } = useI18n();
const { icons } = useIcons();
const { getById } = useDataUtils();
const { state: folderState, closeCreateModal } = useFolder();

const folderName = ref('');

const isOpen = computed(() => folderState.value.createModal.isOpen);

const isValid = computed(() => folderName.value.trim().length > 0);

async function handleSubmit() {
  const title = folderName.value.trim();
  if (!title) return;

  const parentId = folderState.value.createModal.parentFolderId;

  if (parentId) {
    const folder = getById(parentId, { types: ItemType.FOLDER });
    if (folder) {
      await $hooks.callHook('data:create:folder:in-folder', {
        folder,
        title,
      });
    }
  } else {
    await $hooks.callHook('data:create:folder', { title });
  }

  handleClose();
}

function handleCancel() {
  handleClose();
}

function handleClose() {
  folderName.value = '';
  closeCreateModal();
}
</script>

<template>
  <UModal
    :open="isOpen"
    :title="t('modal.create_folder.title')"
    @update:open="
      (val: boolean) => {
        if (!val) handleCancel();
      }
    "
  >
    <template #body>
      <form @submit.prevent="handleSubmit">
        <UInput
          v-model="folderName"
          autofocus
          :placeholder="t('placeholder.folder_untitled')"
          :icon="icons.folder"
          size="lg"
        />
      </form>
    </template>

    <template #footer>
      <div class="flex items-center gap-2 w-full justify-end">
        <UButton color="neutral" variant="ghost" @click="handleCancel">
          {{ t('modal.create_folder.cancel') }}
        </UButton>

        <UButton color="primary" :disabled="!isValid" @click="handleSubmit">
          {{ t('modal.create_folder.submit') }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>
