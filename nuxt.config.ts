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
    dirs: ['composables/**', 'types/*', 'app/types'],
  },
  modules: [
    '@nuxt/image',
    '@nuxt/ui',
    '@nuxt/eslint',
    '@nuxtjs/i18n',
    '@pinia/nuxt',
  ],

  // UI
  css: ['~/assets/css/main.css'],
  colorMode: {
    preference: 'system',
    fallback: 'light',
  },
  ui: {
    theme: {
      colors: ['plain'],
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

  // Custom (public = exposé au client)
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
