<script setup lang="ts">
import type { BreadcrumbItem, DropdownMenuItem } from '@nuxt/ui';

interface Props {
  ancestors: string[];
  maxVisibleItems?: number;
}

const props = withDefaults(defineProps<Props>(), {
  maxVisibleItems: 3,
});

const store = useDataStore();
const route = useRoute();

const resolvedAncestors = computed(() => {
  return props.ancestors
    .filter((id) => id !== 'root')
    .map((id) => {
      const item = store.map.get(id);
      if (!item || item.type !== ItemType.FOLDER) return undefined;
      return { id: item.id, label: item.title, icon: 'i-lucide-folder' };
    })
    .filter(
      (f): f is { id: string; label: string; icon: string } => f !== undefined
    );
});

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  const currentRouteId = route.params.id as string | undefined;

  return resolvedAncestors.value.map((item, index, arr) => {
    const isLast = index === arr.length - 1;
    const isCurrentPage = isLast && currentRouteId === item.id;

    return {
      label: item.label,
      icon: item.icon,
      to: isCurrentPage ? undefined : `/${item.id}/`,
      disabled: isCurrentPage,
    };
  });
});

const hiddenItems = computed<DropdownMenuItem[]>(() => {
  const items = breadcrumbs.value;
  const limit = props.maxVisibleItems - 1;

  if (items.length <= props.maxVisibleItems) return [];

  return items.slice(1, -limit).map((item) => ({
    label: item.label,
    icon: item.icon as string,
    to: item.to as string,
  }));
});

const visibleItems = computed<BreadcrumbItem[]>(() => {
  const items = breadcrumbs.value;
  const first = items.at(0);
  const last = items.at(-1);

  if (items.length <= props.maxVisibleItems || !first || !last) {
    return items;
  }

  return [
    first,
    { slot: 'ellipsis' as const, icon: 'i-lucide-ellipsis' },
    last,
  ];
});
</script>

<template>
  <UBreadcrumb
    v-if="resolvedAncestors.length"
    :items="visibleItems"
    class="select-none"
    :ui="{ root: 'text-xs', linkLabel: 'text-xs' }"
  >
    <template #ellipsis>
      <UDropdownMenu :items="hiddenItems">
        <UButton
          icon="i-lucide-ellipsis"
          color="neutral"
          variant="link"
          size="xs"
          class="p-0.5"
        />
      </UDropdownMenu>
    </template>
  </UBreadcrumb>
</template>
