declare module 'nuxt/schema' {
  interface PublicRuntimeConfig {
    notes: {
      title: {
        maxLength: number;
        default: string;
      };
    };
  }
}

export {};
