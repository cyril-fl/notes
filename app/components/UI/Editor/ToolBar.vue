<script setup lang="ts">
import { ToggleGroupItem, ToggleGroupRoot } from 'reka-ui';
/* Define */
const items = useEditorToolBar();

const customize = (
  content: string,
  _name: string | undefined,
  _prefix: string | undefined,
  _provider: string | undefined,
) => {
  return content
    .replace(/stroke-width="[^"]*"/g, `stroke-width="2.5"`) // Change stroke width to 2
    .replace(/animation-duration="[^"]*"/g, `animation-duration="1s"`) // Change animation duration to 1s (for animated icons)
};
/* Data */

/* Methods */
/* Lifecycle */
/* SEO */
</script>

<template>
  <div class="flex gap-2 items-center justify-center">
    <ToggleGroupRoot v-for="(group, g_index) in items" :key="`group-${g_index}`" class="flex bg-elevated items-center gap-2 p-1 border shadow-sm rounded-lg">
      <template v-for="(item, i_index) in group" :key="`item-${g_index}-${i_index}`">
        <ToggleGroupItem :value="item.kind" class="flex items-center">
          <Icon :name="item.icon" :customize="customize" />
        </ToggleGroupItem>
      </template>
  </ToggleGroupRoot>
  </div>
</template>