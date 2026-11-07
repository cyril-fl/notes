<script setup lang="ts">
import type { FetchError } from 'ofetch';
import type { NoteFile, User } from '~/types/items';

const { icons } = useIcons();
const { t } = useI18n();

const api = useApi();
const toast = useToast();
const logout = useLogout();
const baseURL = useRuntimeConfig().public.apiBase;
const input = useTemplateRef('input');

const { data: me } = useLazyAsyncData('me', () => api<User>('/api/me'));

const avatarSrc = computed(() => me.value?.avatar && imageSrc(`/api/files/${me.value.avatar}`, baseURL));
const uploading = ref(false);

// square crops the image to its centered square and scales it to size: a phone photo of several MB becomes a few KB.
// An animated GIF keeps its first frame.
async function square(file: File, size: number): Promise<Blob> {
  const bitmap = await createImageBitmap(file);
  const side = Math.min(bitmap.width, bitmap.height);
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = Math.min(side, size);
  canvas.getContext('2d')!.drawImage(bitmap, (bitmap.width - side) / 2, (bitmap.height - side) / 2, side, side, 0, 0, canvas.width, canvas.height);
  bitmap.close();
  // A browser that cannot encode WebP gives a PNG.
  return new Promise((resolve, reject) => canvas.toBlob((blob) => (blob ? resolve(blob) : reject(new Error('canvas.toBlob'))), 'image/webp', 0.9));
}

async function changeAvatar(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) return;
  uploading.value = true;
  try {
    const image = await square(file, 256);
    const name = `avatar.${image.type === 'image/webp' ? 'webp' : 'png'}`;
    const saved = await api<NoteFile>('/api/files', { method: 'POST', body: image, query: { name } });
    me.value = await api<User>('/api/me', { method: 'PATCH', body: { avatar: saved.id } });
  } finally {
    uploading.value = false;
    input.value!.value = ''; // The same file can be picked again.
  }
}

async function removeAvatar() {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { avatar: '' } });
}

const name = ref('');
watch(() => me.value?.name, (current) => (name.value = current ?? ''), { immediate: true });

async function rename() {
  me.value = await api<User>('/api/me', { method: 'PATCH', body: { name: name.value }, conflict: t('account.nameTaken') });
  toast.add({ title: t('account.renamed'), color: 'success' });
}

const current = ref('');
const next = ref('');

async function changePassword() {
  try {
    await api('/api/me/password', { method: 'PATCH', body: { current: current.value, new: next.value }, quiet: true });
  } catch (error) {
    if ((error as FetchError).status === 429) {
      toast.add({ title: t('api.tooManyTries', { wait: retryWait(error as FetchError) }), color: 'error' });
      return;
    }
    // The API's messages are in English: the two the user can fix are said here.
    const message: string = (error as FetchError).data?.error ?? '';
    const key = message.includes('current password') ? 'account.wrongPassword' : message.includes('characters') ? 'account.shortPassword' : 'api.failed';
    toast.add({ title: t(key), color: 'error' });
    return;
  }
  current.value = next.value = '';
  toast.add({ title: t('account.passwordChanged'), color: 'success' });
}
</script>

<template>
  <div class="p-8">
    <PageHeader :title="$t('account.title')" />

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('account.profile') }}</h2>
      <ul>
        <li class="flex items-center gap-4 border-b border-muted py-3">
          <UAvatar :src="avatarSrc || undefined" :alt="me?.name" size="3xl" />
          <div class="min-w-0 flex-1">
            <p class="text-sm">{{ $t('account.avatar') }}</p>
            <p class="text-xs text-muted">{{ $t('account.avatarDescription') }}</p>
          </div>
          <input ref="input" type="file" accept="image/png,image/jpeg,image/gif,image/webp" class="hidden" @change="changeAvatar">
          <UButton v-if="me?.avatar" :label="$t('account.removeAvatar')" color="neutral" variant="ghost" @click="removeAvatar" />
          <UButton :label="$t('account.chooseAvatar')" :icon="icons.avatar" color="neutral" variant="outline" :loading="uploading" @click="input?.click()" />
        </li>
        <li class="py-3">
          <form class="flex items-center gap-4" @submit.prevent="rename">
            <div class="min-w-0 flex-1">
              <p class="text-sm">{{ $t('account.name') }}</p>
              <p class="text-xs text-muted">{{ $t('account.nameDescription') }}</p>
            </div>
            <UInput v-model="name" autocomplete="username" :aria-label="$t('account.name')" class="w-56" />
            <UButton type="submit" :label="$t('account.save')" color="neutral" variant="outline" :disabled="!name.trim() || name.trim() === me?.name" />
          </form>
        </li>
      </ul>
    </section>

    <section class="mb-8">
      <h2 class="mb-1 text-sm font-semibold">{{ $t('account.password') }}</h2>
      <p class="mb-3 text-xs text-muted">{{ $t('account.passwordDescription') }}</p>
      <form class="flex flex-wrap items-end gap-4" @submit.prevent="changePassword">
        <UFormField :label="$t('account.currentPassword')">
          <UInput v-model="current" type="password" autocomplete="current-password" class="w-56" />
        </UFormField>
        <UFormField :label="$t('account.newPassword')">
          <UInput v-model="next" type="password" autocomplete="new-password" class="w-56" />
        </UFormField>
        <UButton type="submit" :label="$t('account.changePassword')" color="neutral" variant="outline" :disabled="!current || !next" />
      </form>
    </section>

    <section>
      <h2 class="mb-1 text-sm font-semibold">{{ $t('account.session') }}</h2>
      <div class="flex items-center gap-4 py-3">
        <p class="min-w-0 flex-1 truncate text-sm">
          {{ $t('settings.loggedInAs') }} <span class="font-medium">{{ me?.name }}</span>
        </p>
        <UButton :label="$t('settings.logout')" :icon="icons.logout" color="neutral" variant="outline" @click="logout" />
      </div>
    </section>
  </div>
</template>
