<script setup lang="ts">
import type { NavigationTreeProps } from '~/types/ui';

const props = defineProps<NavigationTreeProps>();
const { onFolderNavOpen } = useUIEvents();
const { getListStateIcon } = useIcons();

const isOpen = ref(false);

async function handleOpen(id: string) {
  const isTarget = id === props.item.id;
  console.log(
    'handleOpen called with id:',
    id,
    'isTarget:',
    isTarget,
    'isOpen:',
    isOpen.value
  );

  if (id !== props.item.id || isOpen.value) {
    await nextTick();
    return;
  }

  console.log('Opening folder:', props.item.id);

  isOpen.value = true;
}

onMounted(() => {
  const dispose = onFolderNavOpen(handleOpen);
  onBeforeUnmount(dispose);
});
</script>

<template>
  <div class="space-y-1">
    <template v-if="!item.to">
      <p class="text-sm font-semibold text-muted px-2 select-none">
        {{ $t(item.label) }}
      </p>
      <slot />
    </template>
    <template v-else>
      <UCollapsible v-model:open="isOpen">
        <template #default="{ open }">
          <div class="flex items-center select-none">
            <UINavigationTreeMenuItem :item="item" />
            <UIcon
              v-if="item.children?.length"
              :name="getListStateIcon(open)"
              class="ml-2"
            />
          </div>
        </template>

        <template #content>
          <div class="space-y-1 pl-4">
            <slot />
          </div>
        </template>
      </UCollapsible>
    </template>
  </div>
</template>

<style scoped></style>
