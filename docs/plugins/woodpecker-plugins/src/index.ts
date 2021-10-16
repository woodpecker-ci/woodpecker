import {
  LoadContext,
  Plugin,
  PluginContentLoadedActions,
} from '@docusaurus/types';
import { Octokit } from '@octokit/rest';
import { components as OctokitComponents } from '@octokit/openapi-types';
import path from 'path';
import { Content, WoodpeckerPlugin, WoodpeckerPluginHeader } from './types';
import * as markdown from './markdown';
import routes from '@generated/routes';

async function loadContent(): Promise<Content> {
  const octokit = new Octokit();

  const codeResults = (
    await octokit.rest.search.code({
      // search for repos in woodpecker-ci org with a file docs.md containing the string WOODPECKER_PLUGIN_DOCS
      q: 'org:woodpecker-ci filename:docs.md WOODPECKER_PLUGIN_DOCS',
    })
  ).data.items;

  const plugins = await Promise.all(
    codeResults
      .filter((i) => i.repository.name.startsWith('plugin-'))
      .map(async (i) => {
        const docsResult = (
          await octokit.repos.getContent({
            owner: 'woodpecker-ci',
            repo: i.repository.name,
            path: '/docs.md',
          })
        ).data as OctokitComponents['schemas']['content-file'];

        const docs = Buffer.from(docsResult.content, 'base64').toString(
          'ascii'
        );

        const header = markdown.getHeader<WoodpeckerPluginHeader>(docs);

        const plugin: WoodpeckerPlugin = {
          name: header?.name || i.repository.name,
          repoName: i.repository.name,
          url: i.repository.html_url,
          icon: header?.icon,
          description: header?.description,
          docs: markdown.getContent(docs),
        };

        return plugin;
      })
  );

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

  const pluginsJsonPath = await createData(
    'plugins.json',
    JSON.stringify(plugins)
  );

  await Promise.all(
    plugins.map(async (plugin) => {
      const pluginJsonPath = await createData(
        `plugin-${plugin.repoName}.json`,
        JSON.stringify(plugin)
      );

      addRoute({
        path: `/plugins/${plugin.repoName}`,
        component: '@theme/WoodpeckerPlugin',
        modules: {
          plugin: pluginJsonPath,
        },
        exact: true,
      });
    })
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

export default function pluginWoodpeckerPluginsIndex(
  context: LoadContext,
  options: any
): Plugin<Content> {
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
      return [path.join(__dirname, '..', 'dist', '**', '*.{js,jsx}')];
    },
  };
}

const getSwizzleComponentList = (): string[] => {
  return ['WoodpeckerPluginList', 'WoodpeckerPlugin'];
};

export { getSwizzleComponentList };
