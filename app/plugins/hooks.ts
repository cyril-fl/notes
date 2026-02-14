import { defineNuxtPlugin } from '#app';
import { Hookable } from 'hookable';

export default defineNuxtPlugin({
  name: 'hooks',
  setup() {
    const hooks = new Hookable();

    return {
      provide: {
        hooks,
      },
    };
  },
});
