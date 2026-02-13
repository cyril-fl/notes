// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from '@tailwindcss/vite';

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  typescript: {
    // NOTE: help resolve await errors in .vue files
    tsConfig: {
      compilerOptions: {
        target: 'ES2022',
        module: 'ESNext',
      },
      include: ['types/**/*.d.ts'],
    },
  },
  // Directories
  imports: {
    dirs: ['composables/**/*', 'types/**/*', 'app/types'],
  },
  modules: [
    '@nuxt/eslint',
    '@nuxtjs/i18n',
    '@pinia/nuxt',
    '@nuxt/image',
    '@nuxt/ui',
    '@nuxt/icon',
    '@nuxtjs/google-fonts',
  ],

  // UI
  css: ['~/assets/css/main.css'],
  colorMode: {
    preference: 'system',
    fallback: 'light',
    storageKey: 'nuxt-color-mode',
  },
  icon: {
    mode: 'css',
    cssLayer: 'base',
  },

  googleFonts: {
    families: {
      'Reddit Sans': true,
    },
  },

  // Utilities
  i18n: {
    strategy: 'no_prefix',
    defaultLocale: 'fr',
    locales: [{ code: 'fr', language: 'fr-FR', file: 'fr-FR.json' }],
  },

  // Server
  vite: {
    plugins: [tailwindcss()],
  },
  nitro: {
    imports: {},
    storage: {
      mongodb: {
        driver: 'mongodb',
        connectionString:
          process.env.MONGODB_URI || 'mongodb://localhost:27017/',
        databaseName: 'notes-pad',
        collectionName: 'items',
      },
    },
  },

  // Config
  runtimeConfig: {
    public: {
      notes: {
        title: {
          maxLength: 50,
          default: 'notes.default_title',
        },
      },
    },
  },
});
