import { LoadContext, Plugin, PluginContentLoadedActions } from '@docusaurus/types';
import path from 'path';
import fs from 'fs';
import axios, { AxiosError } from 'axios';
import { Content, WoodpeckerPlugin, WoodpeckerPluginHeader, WoodpeckerPluginIndexEntry } from './types';
import * as markdown from './markdown';

async function loadContent(): Promise<Content> {
  const file = path.join(__dirname, '..', 'plugins.json');

  const pluginsIndex = JSON.parse(fs.readFileSync(file).toString()) as { plugins: WoodpeckerPluginIndexEntry[] };

  const plugins = (
    await Promise.all(
      pluginsIndex.plugins.map(async (i) => {
        let docsContent: string;
        try {
          const response = await axios(i.docs);
          docsContent = response.data;
        } catch (e) {
          console.error("Can't fetch docs file", i.docs, (e as AxiosError).message);
          return undefined;
        }

        const docsHeader = markdown.getHeader<WoodpeckerPluginHeader>(docsContent);
        const docsBody = markdown.getContent(docsContent);

        if (!docsHeader.name) {
          return undefined;
        }

        return <WoodpeckerPlugin>{
          name: docsHeader.name || i.name,
          url: docsHeader.url,
          icon: docsHeader.icon,
          description: docsHeader.description,
          docs: docsBody,
          tags: docsHeader.tags || [],
          author: docsHeader.author,
          containerImage: docsHeader.containerImage,
          containerImageUrl: docsHeader.containerImageUrl,
          verified: i.verified || false,
        };
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
