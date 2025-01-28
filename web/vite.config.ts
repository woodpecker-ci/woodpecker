import { readdirSync } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite';
import vue from '@vitejs/plugin-vue';
import type { Plugin } from 'vite';
import prismjs from 'vite-plugin-prismjs';
import svgLoader from 'vite-svg-loader';
import type { ViteUserConfig } from 'vitest/config';
import { defineConfig } from 'vitest/config';

function woodpeckerInfoPlugin(): Plugin {
  return {
    name: 'woodpecker-info',
    configureServer() {
      const info =
        '1) Please add `WOODPECKER_DEV_WWW_PROXY=http://localhost:8010` to your `.env` file.\n' +
        'After starting the woodpecker server as well you should now be able to access the UI at http://localhost:8000/\n\n' +
        '2) If you want to run the vite dev server (`pnpm start`) within a container please set `VITE_DEV_SERVER_HOST=0.0.0.0`.';
      console.log(info);
    },
  };
}

function externalCSSPlugin(): Plugin {
  return {
    name: 'external-css',
    transformIndexHtml: {
      order: 'post',
      handler() {
        return [
          {
            tag: 'link',
            attrs: { rel: 'stylesheet', type: 'text/css', href: '/assets/custom.css' },
            injectTo: 'head',
          },
        ];
      },
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    VueI18nPlugin({
      include: path.resolve(__dirname, 'src/assets/locales/**'),
    }),
    (() => {
      const virtualModuleId = 'virtual:vue-i18n-supported-locales';
      const resolvedVirtualModuleId = `\0${virtualModuleId}`;

      const filenames = readdirSync('src/assets/locales/').map((filename) => filename.replace('.json', ''));

      return {
        name: 'vue-i18n-supported-locales',

        resolveId(id) {
          if (id === virtualModuleId) {
            return resolvedVirtualModuleId;
          }
        },

        load(id) {
          if (id === resolvedVirtualModuleId) {
            return `export const SUPPORTED_LOCALES = ${JSON.stringify(filenames)}`;
          }
        },
      };
    })(),
    svgLoader(),
    externalCSSPlugin(),
    woodpeckerInfoPlugin(),
    prismjs({
      languages: ['yaml'],
    }),
  ],
  resolve: {
    alias: {
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
  logLevel: 'warn',
  server: {
    allowedHosts: true,
    host: process.env.VITE_DEV_SERVER_HOST ?? '127.0.0.1',
    port: 8010,
  },
  test: {
    globals: true,
    environment: 'jsdom',
  },
} as ViteUserConfig);
