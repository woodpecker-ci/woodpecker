import { LoadContext, Plugin, PluginContentLoadedActions } from '@docusaurus/types';
import { Octokit } from '@octokit/rest';
import { components as OctokitComponents } from '@octokit/openapi-types';
import path from 'path';
import { Content, WoodpeckerPlugin, WoodpeckerPluginHeader } from './types';
import * as markdown from './markdown';

const octokit = new Octokit();

async function getDocs(repoName: string): Promise<string | undefined> {
  try {
    const docsResult = (
      await octokit.repos.getContent({
        owner: 'woodpecker-ci',
        repo: repoName,
        path: '/docs.md',
      })
    ).data as OctokitComponents['schemas']['content-file'];

    return Buffer.from(docsResult.content, 'base64').toString('ascii');
  } catch (e) {
    console.error("Can't fetch docs file for repository", repoName, e);
  }

  return undefined;
}

async function loadContent(): Promise<Content> {
  const repositories = (
    await octokit.rest.search.repos({
      // search for repos in woodpecker-ci org with the topic: woodpecker-plugin including forks
      q: 'org:woodpecker-ci topic:woodpecker-plugin fork:true',
    })
  ).data.items;

  console.log(repositories.map((r) => r.name));

  const plugins = (
    await Promise.all(
      repositories.map(async (repo) => {
        const docs = await getDocs(repo.name);
        if (!docs) {
          return undefined;
        }

        const header = markdown.getHeader<WoodpeckerPluginHeader>(docs);
        const body = markdown.getContent(docs);

        const plugin: WoodpeckerPlugin = {
          name: header?.name || repo.name,
          repoName: repo.name,
          url: repo.html_url,
          icon: header?.icon,
          description: header?.description,
          docs: body,
        };

        return plugin;
      }),
    )
  ).filter((plugin) => plugin);

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
    plugins.map(async (plugin) => {
      const pluginJsonPath = await createData(`plugin-${plugin.repoName}.json`, JSON.stringify(plugin));

      addRoute({
        path: `/plugins/${plugin.repoName}`,
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
      return [path.join(__dirname, '..', 'dist', '**', '*.{js,jsx}')];
    },
  };
}

const getSwizzleComponentList = (): string[] => {
  return ['WoodpeckerPluginList', 'WoodpeckerPlugin'];
};

export { getSwizzleComponentList };
