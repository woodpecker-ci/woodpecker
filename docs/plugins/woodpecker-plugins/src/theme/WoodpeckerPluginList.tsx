import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import { WoodpeckerPlugin } from '../types';

function PluginPanel({ plugin }: { plugin: WoodpeckerPlugin }) {
  const pluginUrl = `/plugins/${plugin.name}`;
  return (
    <div className={clsx('col col--6')}>
      <div className={clsx('card margin-horiz--sm margin-vert--md ')}>
        <div className={clsx('card__header row')}>
          <div className={clsx('col col--8')}>
            <a href={pluginUrl}>
              <h3>{plugin.name}</h3>
            </a>
            <p>{plugin.description}</p>
          </div>
          <a href={pluginUrl} className={clsx('col col--4 text--right')}>
            <img src={plugin.icon} width="100" height="100" />
          </a>
        </div>
        <div className={clsx('card__footer')}>
          <a href={pluginUrl} className={clsx('button button--secondary button--outline button--block ')}>
            Open {plugin.name}
          </a>
        </div>
      </div>
    </div>
  );
}

export function WoodpeckerPluginList({ plugins }: { plugins: WoodpeckerPlugin[] }) {
  const applyForIndexUrl =
    'https://github.com/woodpecker-ci/woodpecker/edit/master/docs/plugins/woodpecker-plugins/plugins.json';

  const NewPluginPanel = () => (
    <div className={clsx('col col--6')}>
      <div className={clsx('card margin-horiz--sm margin-vert--md ')}>
        <div className={clsx('card__header row')}>
          <div className={clsx('col col--8')}>
            <a href={applyForIndexUrl}>
              <h3>Add your own plugin</h3>
            </a>
            <p>You can simply add your own plugin to this index.</p>
          </div>
          <a
            href={applyForIndexUrl}
            target="_blank"
            rel="noopener noreferrer"
            className={clsx('col col--4 text--right')}
          >
            <svg width="100" height="100" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M88.2357 38.0952H61.9048V11.7643C61.9048 5.29524 56.4714 0 50 0C43.5286 0 38.0952 5.29524 38.0952 11.7643V38.0952H11.7643C5.29524 38.0952 0 43.5286 0 50C0 56.4714 5.29524 61.9048 11.7643 61.9048H38.0952V88.2357C38.0952 94.7048 43.5286 100 50 100C56.4714 100 61.9048 94.7048 61.9048 88.2357V61.9048H88.2357C94.7048 61.9048 100 56.4714 100 50C100 43.5286 94.7048 38.0952 88.2357 38.0952Z"
                fill="#4CAF50"
              />
            </svg>
          </a>
        </div>
        <div className={clsx('card__footer')}>
          <a href={applyForIndexUrl} className={clsx('button button--secondary button--outline button--block ')}>
            Add your own plugin
          </a>
        </div>
      </div>
    </div>
  );

  return (
    <Layout title="Woodpecker CI plugins" description="List of all Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div className="container">
            <div className="row">
              <NewPluginPanel />
              {plugins.map((plugin, idx) => (
                <PluginPanel key={idx} plugin={plugin} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default WoodpeckerPluginList;
