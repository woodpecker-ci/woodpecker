import fs from 'fs';
import path from 'path';
import { LoadContext, Plugin, PluginContentLoadedActions } from '@docusaurus/types';
import axios, { AxiosError } from 'axios';

import * as markdown from './markdown';
import { Content, WoodpeckerPlugin, WoodpeckerPluginHeader, WoodpeckerPluginIndexEntry } from './types';

async function loadContent(): Promise<Content> {
  const file = path.join(__dirname, '..', 'plugins.json');

  const pluginsIndex = JSON.parse(fs.readFileSync(file).toString()) as { plugins: WoodpeckerPluginIndexEntry[] };

  const plugins = (
    await Promise.all(
      pluginsIndex.plugins.map(async (i): Promise<WoodpeckerPlugin | undefined> => {
        let docsContent: string;
        try {
          const response = await axios(i.docs);
          docsContent = response.data;
        } catch (e) {
          console.error("Can't fetch docs file", i.docs, (e as AxiosError).message);
          return undefined;
        }

        let docsHeader: WoodpeckerPluginHeader;
        try {
          docsHeader = markdown.getHeader<WoodpeckerPluginHeader>(docsContent);
        } catch (e) {
          console.error("Can't get header from docs file", i.docs, (e as Error).message);
          return undefined;
        }

        const docsBody = markdown.getContent(docsContent);

        if (!docsHeader.name) {
          return undefined;
        }

        let pluginIconDataUrl: string | undefined;
        if (docsHeader.icon) {
          try {
            const response = await axios(docsHeader.icon, {
              responseType: 'arraybuffer',
            });
            pluginIconDataUrl = `data:${response.headers['content-type'].toString()};base64,${Buffer.from(
              response.data,
              'binary',
            ).toString('base64')}`;
          } catch (e) {
            console.error("Can't fetch plugin icon", docsHeader.icon, (e as AxiosError).message);
          }
        }

        return {
          name: docsHeader.name,
          url: docsHeader.url,
          icon: docsHeader.icon,
          description: docsHeader.description,
          docs: docsBody,
          tags: docsHeader.tags || [],
          author: docsHeader.author,
          containerImage: docsHeader.containerImage,
          containerImageUrl: docsHeader.containerImageUrl,
          verified: i.verified || false,
          iconDataUrl: pluginIconDataUrl,
        } satisfies WoodpeckerPlugin;
      }),
    )
  ).filter<WoodpeckerPlugin>((plugin): plugin is WoodpeckerPlugin => plugin !== undefined);

  return {
    plugins,
  };
}

async function contentLoaded({
  content: { plugins },
  actions,
}: {
  content: Content;
  actions: PluginContentLoadedActions;
}): Promise<void> {
  const { createData, addRoute } = actions;

  const pluginsJsonPath = await createData('plugins.json', JSON.stringify(plugins));

  await Promise.all(
    plugins.map(async (plugin, i) => {
      const pluginJsonPath = await createData(`plugin-${i}.json`, JSON.stringify(plugin));

      addRoute({
        path: `/plugins/${plugin.name}`,
        component: '@theme/WoodpeckerPlugin',
        modules: {
          plugin: pluginJsonPath,
        },
        exact: true,
      });
    }),
  );

  addRoute({
    path: '/plugins',
    component: '@theme/WoodpeckerPluginList',
    modules: {
      plugins: pluginsJsonPath,
    },
    exact: true,
  });
}

export default function pluginWoodpeckerPluginsIndex(context: LoadContext, options: any): Plugin<Content> {
  return {
    name: 'woodpecker-plugins',
    loadContent,
    contentLoaded,
    getThemePath() {
      return path.join(__dirname, '..', 'dist', 'theme');
    },
    getTypeScriptThemePath() {
      return path.join(__dirname, '..', 'src', 'theme');
    },
    getPathsToWatch() {
      return [path.join(__dirname, '..', 'dist', '**', '*.{js,jsx,css}')];
    },
  };
}

const getSwizzleComponentList = (): string[] => {
  return ['WoodpeckerPluginList', 'WoodpeckerPlugin'];
};

export { getSwizzleComponentList };
