<script setup lang="ts">
const props = defineProps<{ id: SettingsSection }>();
const route = useRoute();
// On the section's own page, the page header already names it.
const alone = computed(() => route.params.section === props.id);
</script>

<template>
  <div v-if="!alone" class="mb-3">
    <h2 class="text-base font-semibold">{{ $t(`settings.sections.${id}`) }}</h2>
    <p class="text-xs text-muted">{{ $t(`settings.sections.${id}Description`) }}</p>
  </div>
  <p v-else class="mb-3 text-xs text-muted">{{ $t(`settings.sections.${id}Description`) }}</p>
</template>
