<script setup lang="ts">
import type { NuxtError } from '#app';

const { icons } = useIcons();

defineProps<{
  error: NuxtError;
}>();

function handleBack() {
  if (window.history.length > 1) {
    clearError();
    window.history.back();
  } else {
    clearError({ redirect: NAVIGATION.home });
  }
}
</script>

<template>
  <NuxtLayout name="pages">
    <UIPageSection class="items-center justify-center">
      <UIcon :name="icons.notFound" class="size-16 text-muted" />
      <hgroup class="text-center space-y-6">
        <h1 class="text-4xl font-bold text-default">
          {{ $t('pages.error.title') }}
        </h1>
        <h3 class="text-lg text-muted max-w-md mx-auto">
          {{ $t('pages.error.description') }}
        </h3>
      </hgroup>
      <div class="flex gap-4 justify-center pt-4">
        <UButton
          :label="$t('button.back')"
          color="neutral"
          variant="soft"
          size="xl"
          @click="handleBack"
        />
        <UButton
          :label="$t('button.go_home')"
          color="primary"
          size="xl"
          @click="clearError({ redirect: NAVIGATION.home })"
        />
      </div>
    </UIPageSection>
  </NuxtLayout>
</template>
