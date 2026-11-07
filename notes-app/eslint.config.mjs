// @ts-check
import prettier from 'eslint-config-prettier';
import withNuxt from './.nuxt/eslint.config.mjs';

// Prettier last: it turns off the rules that only fight over formatting.
export default withNuxt(prettier);
