<script setup lang="ts">
// Imports
import {
  EditableArea,
  EditableInput,
  EditablePreview,
  EditableRoot,
} from 'reka-ui';

import type { ContextMenuItem } from '@nuxt/ui';
import type { NavigationTreeProps } from '~/types/ui';

// Define
const { $hooks } = useNuxtApp();
const props = defineProps<NavigationTreeProps>();

const model = ref(props.item.label);
const editable = useTemplateRef('editable');
const inputRef = useTemplateRef('inputRef');
const { getById } = useDataUtils();
const { t } = useI18n();

const items = ref<ContextMenuItem[][]>([
    [
      {
        label: t('menu.context.add_folder'),
        icon: 'i-lucide-folder-plus',
        onSelect: async () => {
          const folder = getById(props.item.id, { types: ItemType.FOLDER });
          if (!folder) return;

          const results = await $hooks.callHookParallel(
            'data:create:folder:in-folder',
            { folder }
          );

          const newFolder = Array.isArray(results)
            ? results.find((r) => r !== undefined)
            : undefined;
          if (!newFolder) return;

          $hooks.callHook('navigation:folder:blur-except', props.item.id);
          $hooks.callHook('navigation:folder:edit', newFolder.id);
        },
      },
      {
        label: t('menu.context.add_note'),
        icon: 'i-lucide-file-plus',
        to: `${NAVIGATION.newNote}${props.item.id}/`,
      },
    ],
    [
      {
        label: t('menu.context.rename'),
        icon: 'i-lucide-pencil',
        onSelect: () => {
          $hooks.callHook('navigation:folder:edit', props.item.id);
        },
        disabled: props.item.id === 'root',
      },
      {
        label: t('menu.context.delete'),
        icon: 'i-lucide-trash',
        color: 'error',
        onSelect: () => {
          $hooks.callHook('data:delete:id', props.item.id);
        },
        disabled: props.item.id === 'root',
      },
    ],
]);
// Data

// Methods
async function handleEdit(id: string) {
  if (id !== props.item.id) return;
  editable.value?.edit();
  await nextTick();
  setTimeout(() => {
    const input = inputRef.value;
    if (input) {
      input.focus();
      input.selectionStart = input.selectionEnd = input.value.length;
    }
  }, 50);
}

function handleSubmit(value: string | null | undefined) {
  if (!value) return;

  const trimmedValue = value.trim();
  if (!trimmedValue || trimmedValue === props.item.label) return;

  $hooks.callHook('data:update', props.item.id, { title: trimmedValue });
}

function handleBlur(id: string) {
  if (id !== props.item.id) return;
  editable.value?.submit();
}

function handleBlurExcept(id: string) {
  if (id === props.item.id) return;
  editable.value?.submit();
}

// Lifecycle
onMounted(() => {
  $hooks.addHooks({
    'navigation:folder:edit': handleEdit,
    'navigation:folder:blur': handleBlur,
    'navigation:folder:blur-except': handleBlurExcept,
  });
});
// SEO
</script>

<template>
  <li>
    <UContextMenu :items="items">
      <UButton
        :icon="item.icon"
        :to="item.to"
        class="w-full justify-start"
        color="neutral"
        variant="ghost"
      >
        <EditableRoot
          ref="editable"
          v-slot="{ isEditing, submit }"
          v-model="model"
          placeholder="Enter text..."
          activation-mode="none"
          auto-resize
          @submit="handleSubmit"
        >
          <EditableArea
            :class="[isEditing ? 'bg-accented px-1 py-0.5' : '']"
            class="rounded-md"
            @keydown.enter.prevent="submit"
          >
            <EditablePreview />
            <EditableInput as-child>
              <input ref="inputRef" type="text" @click.stop />
            </EditableInput>
          </EditableArea>
        </EditableRoot>
      </UButton>
    </UContextMenu>
  </li>
</template>
