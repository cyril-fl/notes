<script setup lang="ts">
import type { FetchError } from 'ofetch';

// No sidebar: there is nothing to show before logging in.
definePageMeta({ layout: false });

const api = useApi();
const route = useRoute();
const { t } = useI18n();

const name = ref('');
const password = ref('');
const failure = ref('');
const pending = ref(false);

async function login() {
  pending.value = true;
  failure.value = '';
  try {
    await api('/api/login', { method: 'POST', body: { name: name.value, password: password.value }, quiet: true });
    // What was read without a session failed: the pages read it again.
    clearNuxtData();
    await navigateTo((route.query.redirect as string | undefined) ?? '/', { replace: true });
  } catch (error) {
    const status = (error as FetchError).status;
    failure.value =
      status === 401
        ? t('login.badCredentials')
        : status === 429
          ? t('api.tooManyTries', { wait: retryWait(error as FetchError) })
          : t('api.down');
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <main class="flex min-h-dvh items-center justify-center bg-muted p-4">
    <form class="w-full max-w-xs space-y-4 rounded-xl bg-default p-6" @submit.prevent="login">
      <h1 class="text-xl font-semibold tracking-tight">{{ $t('common.notes') }}</h1>
      <UFormField :label="$t('login.name')">
        <UInput v-model="name" autocomplete="username" autofocus class="w-full" />
      </UFormField>
      <UFormField :label="$t('login.password')">
        <UInput v-model="password" type="password" autocomplete="current-password" class="w-full" />
      </UFormField>
      <p v-if="failure" class="text-sm text-error" role="alert">{{ failure }}</p>
      <UButton type="submit" :label="$t('login.submit')" block :loading="pending" :disabled="!name || !password" />
    </form>
  </main>
</template>
