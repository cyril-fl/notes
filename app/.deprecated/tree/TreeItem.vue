<script lang="ts" setup>
import type { NavigationTreeItem } from '~/composables/tree/useNavigationTree';

const props = defineProps<{
  item: NavigationTreeItem;
  level?: number;
}>();

const currentLevel = computed(() => props.level ?? 0);
const isExpanded = ref(props.item.defaultExpanded ?? false);
const hasChildren = computed(
  () => props.item.children && props.item.children.length > 0
);

const toggleExpand = () => {
  isExpanded.value = !isExpanded.value;
};
</script>

<template>
  <li class="space-y-1">
    <div
      :style="{ paddingLeft: `${currentLevel * 12 + 8}px` }"
      class="flex items-center gap-1 p-2 rounded-md hover:bg-muted transition-colors group"
    >
      <button
        v-if="hasChildren"
        class="p-1 hover:bg-accent rounded"
        @click="toggleExpand"
      >
        <span
          :class="[
            isExpanded ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right',
            'w-4 h-4 transition-transform',
          ]"
        />
      </button>
      <span v-else class="w-6" />

      <NuxtLink
        :to="item.to"
        active-class="font-medium"
        class="flex items-center gap-2 flex-1"
      >
        <span v-if="item.icon" :class="item.icon" />
        <span>{{ item.label }}</span>
      </NuxtLink>
    </div>

    <ul v-if="hasChildren && isExpanded" class="space-y-1">
      <NavigationTreeItem
        v-for="child in item.children"
        :key="child.id"
        :item="child"
        :level="currentLevel + 1"
      />
    </ul>
  </li>
</template>
