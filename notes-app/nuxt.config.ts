export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  // The API runs on the loopback and answers the browser directly: no server side here.
  ssr: false,
  modules: ['@nuxt/ui', '@vueuse/nuxt'],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: { apiBase: 'http://localhost:8080' },
  },
});
