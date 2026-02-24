<script setup lang="ts">
import {
  ToolbarRoot,
  ToolbarSeparator,
  ToolbarToggleGroup,
  ToolbarToggleItem,
} from 'reka-ui';

const items = useEditorToolBar();
const { getListStateIcon } = useIcons();
const isOpen = ref(false);
</script>

<template>
  <ToolbarRoot
    orientation="horizontal"
    :loop="true"
    class="flex gap-2 items-center shrink-0"
  >
    <div v-show="isOpen" class="flex gap-2 items-center overflow-x-auto">
      <template v-for="(group, g_index) in items" :key="`group-${g_index}`">
        <ToolbarSeparator
          v-if="g_index > 0"
          class="w-px h-5 bg-muted shrink-0"
        />
        <ToolbarToggleGroup
          type="multiple"
          class="flex bg-elevated items-center gap-1 border-muted shadow-sm rounded-lg overflow-hidden shrink-0"
        >
          <ToolbarToggleItem
            v-for="(item, i_index) in group"
            :key="`item-${g_index}-${i_index}`"
            :value="item.kind"
            class="flex items-center hover:bg-accented/30 active:bg-accented/50 p-1 active:text-dimmed"
            @click="item.onClick"
          >
            <UIcon :name="item.icon" class="text-lg font-bold" />
          </ToolbarToggleItem>
        </ToolbarToggleGroup>
      </template>
    </div>
    <UButton
      :icon="getListStateIcon(isOpen, 'horizontal')"
      color="neutral"
      variant="soft"
      size="xs"
      square
      @click="isOpen = !isOpen"
    />
  </ToolbarRoot>
</template>
