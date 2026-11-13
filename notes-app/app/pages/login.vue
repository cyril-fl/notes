<script setup lang="ts">
import type { FetchError } from 'ofetch';
import type { AddUserMode } from '~/types/items';

// No sidebar: there is nothing to show before logging in.
definePageMeta({ layout: false });

const api = useApi();
const route = useRoute();
const { t } = useI18n();

// What this server allows: a base without an account waits for its first one, so the page hands over to the sign-up
// at once. In the public mode the sign-up is just offered; in the invitation mode it takes an invitation link.
const { data: mode } = await useAsyncData('add-user', () => api<AddUserMode>('/api/add-user', { quiet: true }));
const canSignup = computed(() => mode.value?.allowed && mode.value.mode === 'public');
watchEffect(() => {
  if (mode.value?.allowed && mode.value.mode === 'invitation') navigateTo('/signup', { replace: true });
});

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
      <UButton v-if="canSignup" :label="$t('signup.title')" to="/signup" variant="link" color="neutral" block />
    </form>
  </main>
</template>
