<script setup lang="ts">
const { icons } = useIcons();

definePageMeta({
  layout: 'pages',
});
</script>

<template>
  <UIPageSection class="items-center justify-center">
    <div
      class="bg-primary text-inverted squircle-round p-4 flex items-center justify-center"
    >
      <NuxtImg
        src="/logo/dark-xl.svg"
        alt="Logo"
        width="64"
        height="64"
        class="size-16 mx-auto"
      />
    </div>
    <hgroup class="text-center space-y-6">
      <h1 class="text-4xl font-bold text-default">
        {{ $t('pages.home.title') }}
      </h1>
      <h3 class="text-lg text-muted max-w-md mx-auto">
        {{ $t('pages.home.description') }}
      </h3>
    </hgroup>
    <div class="flex gap-4 justify-center pt-4">
      <UButton
        :label="$t('button.open_graph')"
        color="neutral"
        :icon="icons.branch"
        size="lg"
        :to="NAVIGATION.graph"
        variant="soft"
      />
      <UButton
        :label="$t('button.open_editor')"
        color="primary"
        :icon="icons.note"
        size="lg"
        :to="NAVIGATION.notes"
      />
    </div>
  </UIPageSection>
</template>
