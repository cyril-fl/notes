<script setup lang="ts">
// Imports
import type { NavigationTreeProps } from '~/types/ui';
// Define
const { $hooks } = useNuxtApp();
const props = defineProps<NavigationTreeProps>();
const hookDisposers: Array<() => void> = [];

// Data
const isOpen = ref(false);

// Methods
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

// Lifecycle
onMounted(() => {
  hookDisposers.push($hooks.hook('folder-navigation:open', handleOpen));
});

onBeforeUnmount(() => {
  hookDisposers.forEach((dispose) => dispose());
});

// SEO
</script>

<template>
  <div class="space-y-1">
    <p v-if="!item.to" class="text-sm font-semibold text-muted px-2">
      {{ $t(item.label) }}
    </p>
    <template v-else>
      <!-- <UCollapsible v-model:open="isOpen"> -->
      <!-- <template #default="{ open }"> -->
      <div class="flex items-center">
        <UITreeMenuItem :item="item" />
        <!-- <UIcon
              v-if="item.children?.length"
              :name="getListStateIcon(open)"
              class="ml-2"
            /> -->
      </div>
      <!-- </template> -->

      <!-- <template #content> -->
      <div class="space-y-1 pl-4">
        <slot />
      </div>
      <!-- </template> -->
      <!-- </UCollapsible> -->
    </template>
    <slot v-if="!item.to" />
  </div>
</template>

<style scoped></style>
