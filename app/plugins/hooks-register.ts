import { defineNuxtPlugin } from '#app';

export default defineNuxtPlugin({
  name: 'hooks-register',
  setup() {
    useHooksBuilder().register(createDataApiRegistrar).build();
  },
  hooks: {
    'app:mounted'() {
      console.log('🚀 [PLUGIN] - Hooks registered and ready to use!');
    },
  },
  dependsOn: ['hooks'],
});
