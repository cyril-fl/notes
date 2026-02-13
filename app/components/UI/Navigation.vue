<script lang="ts" setup>
const { links } = useNavigation();
const { folders } = useNavigationTree();
const { handleCreateFolder } = useDataApi();
const { hashtags, hasTags } = useHashtags();
</script>

<template>
  <nav class="flex flex-col gap-2 list-none justify-between">
    <div>
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
    <div>
      <UButton
        icon="i-lucide-folder-plus"
        color="neutral"
        variant="ghost"
        square
        @click="handleCreateFolder('NEW')"
      />
    </div>
  </nav>
</template>
