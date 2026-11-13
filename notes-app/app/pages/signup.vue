<script setup lang="ts">
import type { FetchError } from 'ofetch';
import type { AddUserMode } from '~/types/items';

// No sidebar: there is no account yet.
definePageMeta({
  layout: false,
  // The page is closed unless the server takes a new account: an URL typed by hand goes back to the login.
  middleware: async (to) => {
    const mode = await $fetch<AddUserMode>('/api/add-user', {
      baseURL: useRuntimeConfig().public.apiBase,
      credentials: 'include',
      query: { invite: to.query.invite ?? '' },
    }).catch(() => undefined);
    if (!mode?.allowed) return navigateTo('/login', { replace: true });
  },
});

const api = useApi();
const route = useRoute();
const { t } = useI18n();

const name = ref('');
const password = ref('');
const again = ref('');
const failure = ref('');
const pending = ref(false);
const invite = computed(() => (route.query.invite as string | undefined) ?? '');
// The API asks for the same length; checked here too, so the form says it before sending.
const tooShort = computed(() => password.value.length > 0 && password.value.length < 8);
const differs = computed(() => again.value.length > 0 && again.value !== password.value);

async function signup() {
  pending.value = true;
  failure.value = '';
  try {
    await api('/api/signup', {
      method: 'POST',
      body: { name: name.value, password: password.value, invite: invite.value },
      quiet: true,
    });
    // What was read without a session failed: the pages read it again, as after a login.
    clearNuxtData();
    await navigateTo('/', { replace: true });
  } catch (error) {
    const status = (error as FetchError).status;
    failure.value =
      status === 403
        ? invite.value
          ? t('signup.badInvite')
          : t('signup.closed')
        : status === 409
          ? t('signup.nameTaken')
          : status === 400
            ? t('signup.shortPassword', { count: 8 })
            : t('api.down');
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <main class="flex min-h-dvh items-center justify-center bg-muted p-4">
    <form class="w-full max-w-xs space-y-4 rounded-xl bg-default p-6" @submit.prevent="signup">
      <h1 class="text-xl font-semibold tracking-tight">{{ $t('signup.title') }}</h1>
      <UFormField :label="$t('login.name')">
        <UInput v-model="name" autocomplete="username" autofocus class="w-full" />
      </UFormField>
      <UFormField
        :label="$t('login.password')"
        :error="tooShort ? $t('signup.shortPassword', { count: 8 }) : undefined"
      >
        <UInput v-model="password" type="password" autocomplete="new-password" class="w-full" />
      </UFormField>
      <UFormField :label="$t('signup.passwordAgain')" :error="differs ? $t('signup.passwordsDiffer') : undefined">
        <UInput v-model="again" type="password" autocomplete="new-password" class="w-full" />
      </UFormField>
      <p v-if="failure" class="text-sm text-error" role="alert">{{ failure }}</p>
      <UButton
        type="submit"
        :label="$t('signup.submit')"
        block
        :loading="pending"
        :disabled="!name || tooShort || !password || again !== password"
      />
      <UButton :label="$t('signup.backToLogin')" to="/login" variant="link" color="neutral" block />
    </form>
  </main>
</template>
