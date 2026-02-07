<script setup lang="ts">
/* Define */

/* Data */

/* Methods */

/* Lifecycle */

/* SEO */
definePageMeta({
  layouts: 'pages',
});
</script>

<template>
  <section
    class="flex flex-col size-full items-center justify-center min-h-[70vh] gap-6 px-4 text-center"
  >
    <Icon class="size-16 text-default mx-auto" name="i-lucide-notebook-pen" />
    <hgroup class="text-center space-y-6">
      <h1 class="text-4xl font-bold text-default">
        {{ $t('page.home.title') }}
      </h1>
      <p class="text-lg text-muted max-w-md mx-auto">
        {{ $t('page.home.description') }}
      </p>
    </hgroup>
    <div class="flex gap-4 justify-center pt-4">
      <UIButton
        :label="$t('button.open_graph')"
        color="neutral"
        icon="i-lucide-git-branch"
        size="lg"
        to="/graph"
        variant="soft"
      />
      <UIButton
        :label="$t('button.open_editor')"
        color="primary"
        icon="i-lucide-file-text"
        size="lg"
        to="/editor"
      />
    </div>
  </section>
</template>
