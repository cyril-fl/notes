<script setup lang="ts">
import { en, fr } from '@nuxt/ui/locale';

const { locale } = useI18n();
const uiLocale = computed(() => (locale.value === 'en' ? en : fr));
useHead({ titleTemplate: (title) => (title ? `Notes - ${title}` : 'Notes') });
</script>

<template>
  <UApp :locale="uiLocale" :toaster="{ position: 'bottom-right' }">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
