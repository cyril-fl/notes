<script setup lang="ts">
import type { FetchError } from 'ofetch';

// No sidebar: there is nothing to show before logging in.
definePageMeta({ layout: false });

const api = useApi();
const route = useRoute();

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
      status === 401 ? 'Nom ou mot de passe incorrect.' : status === 429 ? 'Trop d’essais ratés : réessayez dans un quart d’heure.' : "L'API ne répond pas.";
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <main class="flex min-h-dvh items-center justify-center bg-muted p-4">
    <form class="w-full max-w-xs space-y-4 rounded-xl bg-default p-6" @submit.prevent="login">
      <h1 class="text-xl font-semibold tracking-tight">Notes</h1>
      <UFormField label="Nom">
        <UInput v-model="name" autocomplete="username" autofocus class="w-full" />
      </UFormField>
      <UFormField label="Mot de passe">
        <UInput v-model="password" type="password" autocomplete="current-password" class="w-full" />
      </UFormField>
      <p v-if="failure" class="text-sm text-error" role="alert">{{ failure }}</p>
      <UButton type="submit" label="Se connecter" block :loading="pending" :disabled="!name || !password" />
    </form>
  </main>
</template>
