// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs';

export default withNuxt(
  {},
  {
    rules: {
      // General
      'no-case-declarations': 'off',
      // Vue
      'vue/multi-word-component-names': 'off',
      'vue/no-multiple-template-root': 'off',
      'vue/no-v-html': 'off',
    },
  }
);
