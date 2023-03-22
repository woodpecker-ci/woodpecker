/* eslint-disable import/no-extraneous-dependencies */
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite';
import vue from '@vitejs/plugin-vue';
import { readdirSync } from 'fs';
import path from 'path';
import IconsResolver from 'unplugin-icons/resolver';
import Icons from 'unplugin-icons/vite';
import Components from 'unplugin-vue-components/vite';
import { defineConfig } from 'vite';
import prismjs from 'vite-plugin-prismjs';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import WindiCSS from 'vite-plugin-windicss';
import svgLoader from 'vite-svg-loader';

function woodpeckerInfoPlugin() {
  return {
    name: 'woodpecker-info',
    configureServer() {
      const info =
        'Please add `WOODPECKER_DEV_WWW_PROXY=http://localhost:8010` to your `.env` file.\n' +
        'After starting the woodpecker server as well you should now be able to access the UI at http://localhost:8000/';
      // eslint-disable-next-line no-console
      console.log(info);
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
        // eslint-disable-next-line consistent-return
        resolveId(id) {
          if (id === virtualModuleId) {
            return resolvedVirtualModuleId;
          }
        },
        // eslint-disable-next-line consistent-return
        load(id) {
          if (id === resolvedVirtualModuleId) {
            return `export const SUPPORTED_LOCALES = ${JSON.stringify(filenames)}`;
          }
        },
      };
    })(),
    WindiCSS(),
    Icons({}),
    svgLoader(),
    Components({
      resolvers: [IconsResolver()],
    }),
    woodpeckerInfoPlugin(),
    prismjs({
      languages: ['yaml'],
    }),
    viteStaticCopy({
      targets: [
        {
          src: 'build.html',
          dest: '',
          rename: 'index.html',
        },
      ],
    }),
  ],
  resolve: {
    alias: {
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
  logLevel: 'warn',
  server: {
    port: 8010,
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: 'src/main.ts',
    },
  },
});
