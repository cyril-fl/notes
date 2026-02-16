<script lang="ts" setup>
const { links } = useNavigation();
const { folders } = useNavigationTree();
const { hashtags, hasTags } = useHashtags();
const { icons } = useIcons();
</script>

<template>
  <nav class="flex h-[calc(100vh-4rem)] flex-col gap-2 list-none">
    <div class="min-h-0 flex-1 overflow-y-auto pr-1">
      <UINavigationMenu
        v-for="(link, index) in links"
        :key="`link-${link.label}-${index}`"
        :item="link"
      />
      <UITreeMenu
        v-for="folder in folders"
        :key="`folder-${folder.label}`"
        :item="folder"
      />

      <template v-if="hasTags">
        <p class="text-sm font-semibold text-muted px-2">
          {{ $t('menu.navigation.hashtags.title') }}
        </p>
        {{ hashtags }}
      </template>
    </div>
    <div class="pt-2">
      <UButton
        :icon="icons.folderadd"
        color="neutral"
        variant="ghost"
        square
        @click="$hooks.callHook('data:create:folder')"
      />
    </div>
  </nav>
</template>
