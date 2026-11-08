<script setup lang="ts">
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';

const props = defineProps(nodeViewProps);
const { icons } = useIcons();

const href = computed<string>(() => props.node.attrs.href);
const host = computed(() => {
  try {
    return new URL(href.value).host;
  } catch {
    return href.value;
  }
});
const large = computed(() => props.node.attrs.size === 'large');
// The browser asks the site itself: no favicon there, the link icon instead.
const noFavicon = ref(false);
watch(host, () => (noFavicon.value = false));
</script>

<template>
  <!-- A click selects the card for the bubble menu: opening it is the menu's job, not the editor's. -->
  <NodeViewWrapper
    as="a"
    :href="href"
    data-link-card
    class="link-card"
    :class="[large ? 'link-card-large' : 'link-card-mini', { 'ring-2 ring-primary/60': selected }]"
    @click.prevent
  >
    <img
      v-if="!noFavicon"
      :src="`https://${host}/favicon.ico`"
      alt=""
      class="link-card-icon"
      referrerpolicy="no-referrer"
      @error="noFavicon = true"
    />
    <UIcon v-else :name="icons.link" class="link-card-icon text-muted" />
    <!-- Mini on one line, large on two with the whole address. -->
    <span class="flex min-w-0" :class="large ? 'flex-col' : 'items-baseline gap-1.5'">
      <span class="truncate font-medium text-default">{{ node.attrs.text || host }}</span>
      <span class="truncate text-xs text-muted">{{ large ? href : host }}</span>
    </span>
  </NodeViewWrapper>
</template>
