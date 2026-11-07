export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  // The API runs on the loopback and answers the browser directly: no server side here.
  ssr: false,
  modules: ['@nuxt/ui', '@vueuse/nuxt', '@nuxtjs/i18n'],
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'fr',
    locales: [
      { code: 'fr', file: 'fr.json' },
      { code: 'en', file: 'en.json' },
    ],
    // The cookie keeps the last language chosen: the login page, without an account, shows it too.
    detectBrowserLanguage: { useCookie: true, cookieKey: 'notes-locale' },
  },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: { apiBase: 'http://localhost:8080' },
  },
});
