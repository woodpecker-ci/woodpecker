/* eslint-disable import/no-extraneous-dependencies */
import vueI18n from '@intlify/vite-plugin-vue-i18n';
import vue from '@vitejs/plugin-vue';
import { readdirSync } from 'fs';
import path from 'path';
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
    vueI18n({
      include: path.resolve(__dirname, 'src/assets/locales/**'),
    }),
    (() => {
      const virtualModuleId = 'virtual:my-module';
      const resolvedVirtualModuleId = `\0${virtualModuleId}`;

      const filenames = readdirSync('src/assets/locales/').map((filename) => filename.replace('.json', ''));

      return {
        name: 'vue-i18n-available-locales',
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
});
