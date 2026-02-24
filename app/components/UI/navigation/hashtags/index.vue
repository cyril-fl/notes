<script setup lang="ts">
const { hashtags, hasTags } = useHashtags();
</script>

<template>
  <div v-if="hasTags">
    <p class="text-sm font-semibold text-muted px-2">
      {{ $t('menu.navigation.hashtags.title') }}
    </p>
    {{ hashtags }}
  </div>
</template>

<style scoped></style>
