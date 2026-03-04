<script setup lang="ts">
const { icons } = useIcons();

const props = defineProps<{
  items: Array<Note | Folder>;
}>();

const firstItem = computed(() => props.items[0]);
const icon = computed(() =>
  firstItem.value?.type === ItemType.FOLDER ? icons.folder : icons.note
);
const title = computed(() => {
  const raw = firstItem.value?.title ?? '';
  return raw.length > 20 ? `${raw.slice(0, 20)}…` : raw;
});
const count = computed(() => props.items.length);
</script>

<template>
  <div
    class="relative flex items-center gap-2 bg-default/80 backdrop-blur-sm rounded-lg shadow-lg px-3 py-2 text-sm max-w-50 pointer-events-none"
  >
    <UIcon :name="icon" class="size-4 shrink-0" />
    <span class="truncate">{{ title }}</span>
    <span
      v-if="count > 1"
      class="absolute -top-2 -right-2 bg-primary text-inverted text-xs font-medium rounded-full size-5 flex items-center justify-center"
    >
      {{ count }}
    </span>
  </div>
</template>
