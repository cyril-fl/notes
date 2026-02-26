<script lang="ts" setup>
const sectionConfig = usePageSection();
</script>

<template>
  <main class="vessel">
    <UICoreSidePanel>
      <UINavigation />
    </UICoreSidePanel>
    <UIPageSection
      :title="sectionConfig.title"
      :description="sectionConfig.description"
      :editable="sectionConfig.editable"
      :searchable="sectionConfig.searchable"
      :context-actions="sectionConfig.contextActions"
      :class="sectionConfig.sectionClass"
      @submit="sectionConfig.onSubmit?.($event)"
      @cancel="sectionConfig.onCancel?.()"
    >
      <slot />
    </UIPageSection>
  </main>
</template>
