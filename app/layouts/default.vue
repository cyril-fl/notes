<script lang="ts" setup>
const sectionConfig = usePageSection();
const { state: confirmState, confirm, cancel } = useConfirmDelete();
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
      :ancestors="sectionConfig.ancestors"
      :class="sectionConfig.sectionClass"
      @submit="sectionConfig.onSubmit?.($event)"
      @cancel="sectionConfig.onCancel?.()"
    >
      <slot />
    </UIPageSection>
    <UIModalConfirmDelete
      :open="confirmState.open"
      :item-type="confirmState.itemType"
      @confirm="confirm"
      @cancel="cancel"
    />
  </main>
</template>
