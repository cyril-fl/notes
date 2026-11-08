<script setup lang="ts">
import { NodeViewWrapper, nodeViewProps } from '@tiptap/vue-3';

const props = defineProps(nodeViewProps);
const { icons } = useIcons();

const href = computed<string>(() => props.node.attrs.href);
const host = computed(() => linkHost(href.value));
const large = computed(() => props.node.attrs.size === 'large');
const mail = computed(() => isMailLink(href.value));
const title = computed(() => props.node.attrs.text || host.value);
// Under the title, unless it says the same: a mail link titled by its address.
const detail = computed(() => (large.value && !mail.value ? href.value : host.value));
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
      v-if="!mail && !noFavicon"
      :src="`https://${host}/favicon.ico`"
      alt=""
      class="link-card-icon"
      referrerpolicy="no-referrer"
      @error="noFavicon = true"
    />
    <UIcon v-else :name="mail ? icons.mail : icons.link" class="link-card-icon text-muted" />
    <!-- Mini on one line, large on two with the whole address. -->
    <span class="flex min-w-0" :class="large ? 'flex-col' : 'items-baseline gap-1.5'">
      <span class="truncate font-medium text-default">{{ title }}</span>
      <span v-if="detail !== title" class="truncate text-xs text-muted">{{ detail }}</span>
    </span>
  </NodeViewWrapper>
</template>
