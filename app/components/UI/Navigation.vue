<script lang="ts" setup>
const { links } = useNavigation();
const { hashtags, hasTags } = useHashtags();
</script>

<template>
  <nav class="flex flex-col gap-2">
    <UINavigationMenu
      v-for="(link, index) in links"
      :key="`link-${link.label}-${index}`"
      :item="link"
    />
    <template v-if="hasTags">
      <p class="text-sm font-semibold text-muted px-2">
        {{ $t('menu.navigation.hashtags.title') }}
      </p>
      {{ hashtags }}
    </template>
  </nav>
</template>
