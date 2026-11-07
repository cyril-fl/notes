// @ts-check
import prettier from 'eslint-config-prettier';
import withNuxt from './.nuxt/eslint.config.mjs';

// Prettier last: it turns off the rules that only fight over formatting.
export default withNuxt(prettier, {
  rules: {
    // Vue 3 has no filters: the rule reads the `|` of a TypeScript union in a template (`as string | undefined`) as one.
    'vue/no-deprecated-filter': 'off',
  },
});
