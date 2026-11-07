<script setup lang="ts">
import colors from 'tailwindcss/colors';
import type { Theme } from '~/utils/themes';

const props = defineProps<{ theme: Theme; active?: boolean }>();
const colorMode = useColorMode();

// A Tailwind palette may be left out of the CSS when no class uses it: its value then comes from tailwindcss/colors, as in Nuxt UI.
function color(name: string, shade: number) {
  const fallback = (colors as unknown as Record<string, Record<number, string>>)[name]?.[shade] ?? '';
  return `var(--color-${name}-${shade}, ${fallback})`;
}

// The shades Nuxt UI takes for the background, the sidebar, the text and the accent, in light and in dark.
const style = computed(() => {
  const dark = colorMode.value === 'dark';
  const { primary, neutral } = props.theme;
  return {
    '--bg': dark ? color(neutral, 900) : '#fff',
    '--side': color(neutral, dark ? 800 : 50),
    '--line': color(neutral, dark ? 600 : 300),
    '--accent': color(primary, dark ? 400 : 600),
  };
});
</script>

<template>
  <button
    type="button"
    class="flex flex-col gap-2 rounded-md p-1 text-left text-sm ring-2 transition-shadow"
    :class="active ? 'ring-primary' : 'ring-transparent hover:ring-accented'"
    :aria-pressed="active"
  >
    <span class="flex h-20 overflow-hidden rounded border border-muted bg-(--bg)" :style="style" aria-hidden="true">
      <span class="flex w-1/3 flex-col gap-1.5 bg-(--side) p-2">
        <span class="h-1.5 w-3/4 rounded-full bg-(--accent)" />
        <span class="h-1.5 w-full rounded-full bg-(--line)" />
        <span class="h-1.5 w-2/3 rounded-full bg-(--line)" />
      </span>
      <span class="flex flex-1 flex-col gap-1.5 p-2">
        <span class="h-2 w-1/2 rounded-full bg-(--line)" />
        <span class="h-1.5 w-full rounded-full bg-(--line) opacity-60" />
        <span class="h-1.5 w-5/6 rounded-full bg-(--line) opacity-60" />
        <span class="mt-auto h-3 w-1/3 self-end rounded-sm bg-(--accent)" />
      </span>
    </span>
    <span class="px-1">{{ $t(`settings.themes.${theme.id}`) }}</span>
  </button>
</template>
