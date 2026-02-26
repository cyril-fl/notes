<script lang="ts" setup>
provideCurrentFolder();

const { icons } = useIcons();
const { createFolder } = useDataActions();

async function onCreateFolder() {
  const result = await createFolder();
  if (!result) return;
  navigateTo(`/${result.id}/?rename=1`);
}
</script>

<template>
  <nav class="flex h-full flex-col space-y-4 list-none">
    <div class="flex gap-2 items-center select-none">
      <div
        class="squircle-round bg-accented gap-2 aspect-square p-[1 px] flex items-center justify-center size-fit"
      >
        <NuxtImg
          src="/logo/dark-sm.svg"
          alt="Logo"
          width="32"
          height="32"
          class="size-8"
        />
      </div>
      <H1 class="text-xl font-bold">{{ $t('app.name') }}</H1>
    </div>
    <div class="min-h-0 flex-1 overflow-y-auto pr-1">
      <UINavigationLinkMenu />
      <UINavigationTreeMenu />
      <UINavigationHashtags />
    </div>
    <div class="pt-2 border-t border-muted">
      <UTooltip :text="$t('pages.new_folder')">
        <UButton
          :icon="icons.folderadd"
          color="neutral"
          variant="soft"
          size="xs"
          class="ml-auto"
          square
          @click="onCreateFolder"
        />
      </UTooltip>

      <UTooltip :text="$t('pages.trash.title')">
        <UButton
          :icon="icons.trash"
          :to="NAVIGATION.trash"
          color="neutral"
          variant="soft"
          size="xs"
          class="ml-auto"
          square
        />
      </UTooltip>
    </div>
  </nav>
</template>
