import { defineConfig } from 'astro/config';

export default defineConfig({
  site: 'https://prestia.web.app',
  i18n: {
    defaultLocale: 'es',
    locales: ['es', 'en', 'pt'],
    routing: {
      prefixDefaultLocale: false,
    },
  },
});
