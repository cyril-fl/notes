import type { ContextMenuItem } from '@nuxt/ui';

export interface PageSectionConfig {
  title?: string;
  description?: string;
  editable?: boolean;
  searchable?: boolean;
  contextActions?: ContextMenuItem[][];
  sectionClass?: string;
  ancestors?: string[];
  onSubmit?: (value: string) => void;
  onCancel?: () => void;
}

const defaultConfig: PageSectionConfig = {
  title: undefined,
  description: undefined,
  editable: false,
  searchable: true,
  contextActions: undefined,
  sectionClass: undefined,
  ancestors: undefined,
  onSubmit: undefined,
  onCancel: undefined,
};

const sectionConfig = shallowRef<PageSectionConfig>(defaultConfig);

export function usePageSection(config?: MaybeRef<PageSectionConfig>) {
  if (config) {
    // Set config immediately
    sectionConfig.value = { ...defaultConfig, ...toValue(config) };

    // Keep in sync if config is reactive
    if (isRef(config)) {
      watch(config, (newConfig) => {
        sectionConfig.value = { ...defaultConfig, ...newConfig };
      });
    }
  }

  return sectionConfig;
}
