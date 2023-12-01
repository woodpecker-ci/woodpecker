/* eslint-disable import/no-extraneous-dependencies */
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite';
import vue from '@vitejs/plugin-vue';
import { copyFile, existsSync, mkdirSync, readdirSync } from 'fs';
import path from 'path';
import replace from 'replace-in-file';
import IconsResolver from 'unplugin-icons/resolver';
import Icons from 'unplugin-icons/vite';
import Components from 'unplugin-vue-components/vite';
import { defineConfig } from 'vite';
import prismjs from 'vite-plugin-prismjs';
import WindiCSS from 'vite-plugin-windicss';
import svgLoader from 'vite-svg-loader';

function woodpeckerInfoPlugin() {
  return {
    name: 'woodpecker-info',
    configureServer() {
      const info =
        '1) Please add `WOODPECKER_DEV_WWW_PROXY=http://localhost:8010` to your `.env` file.\n' +
        'After starting the woodpecker server as well you should now be able to access the UI at http://localhost:8000/\n\n' +
        '2) If you want to run the vite dev server (`pnpm start`) within a container please set `VITE_DEV_SERVER_HOST=0.0.0.0`.';
      // eslint-disable-next-line no-console
      console.log(info);
    },
  };
}

function externalCSSPlugin() {
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

      if (!existsSync('src/assets/dayjsLocales')) {
        mkdirSync('src/assets/dayjsLocales');
      }

      filenames.forEach(async (name) => {
        // English is always directly loaded (compiled by Vite) and thus not copied
        if (name === 'en') {
          return;
        }
        let langName = name;

        // copy dayjs language
        if (name === 'zh-Hans') {
          // zh-Hans is called zh in dayjs
          langName = 'zh';
        } else if (name === 'zh-Hant') {
          // zh-Hant is called zh-cn in dayjs
          langName = 'zh-cn';
        }

        copyFile(
          `node_modules/dayjs/esm/locale/${langName}.js`,
          `src/assets/dayjsLocales/${name}.js`,
          // eslint-disable-next-line promise/prefer-await-to-callbacks
          (err) => {
            if (err) {
              throw err;
            }
          },
        );
      });
      replace.sync({
        files: 'src/assets/dayjsLocales/*.js',
        // remove any dayjs import and any dayjs.locale call
        from: /(?:import dayjs.*'|dayjs\.locale.*);/g,
        to: '',
      });

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
    host: process.env.VITE_DEV_SERVER_HOST || '127.0.0.1',
    port: 8010,
  },
});
