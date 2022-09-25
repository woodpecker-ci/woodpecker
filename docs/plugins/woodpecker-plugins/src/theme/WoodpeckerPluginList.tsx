import React, { useState } from 'react';
import Fuse from 'fuse.js';
import Layout from '@theme/Layout';
import './style.css';
import { WoodpeckerPlugin } from '../types';
import { IconVerified } from './Icons';

function PluginPanel({ plugin }: { plugin: WoodpeckerPlugin }) {
  const pluginUrl = `/plugins/${plugin.name}`;

  return (
    <a href={pluginUrl} className="card shadow--md wp-plugin-card">
      <div className="card__header row">
        <div className="col col--2 text--left">
          <img src={plugin.icon} width="50" height="50" />
        </div>
        <div className="col col--10">
          <h3>{plugin.name}</h3>
          <p>{plugin.description}</p>
          {plugin.tags && (
            <div className="wp-plugin-tags">
              {plugin.tags.map((tag, idx) => (
                <span className="badge badge--success" key={idx}>
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>
      </div>
      {plugin.verified && <div className="wp-plugin-verified">{IconVerified()}</div>}
    </a>
  );
}

export function WoodpeckerPluginList({ plugins }: { plugins: WoodpeckerPlugin[] }) {
  const applyForIndexUrl =
    'https://github.com/woodpecker-ci/woodpecker/edit/master/docs/plugins/woodpecker-plugins/plugins.json';

  const NewPluginPanel = () => (
    <a href={applyForIndexUrl} target="_blank" rel="noopener noreferrer" className="card shadow--md wp-plugin-card">
      <div className="card__header row">
        <div className="col col--2">
          <svg width="50" height="50" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M88.2357 38.0952H61.9048V11.7643C61.9048 5.29524 56.4714 0 50 0C43.5286 0 38.0952 5.29524 38.0952 11.7643V38.0952H11.7643C5.29524 38.0952 0 43.5286 0 50C0 56.4714 5.29524 61.9048 11.7643 61.9048H38.0952V88.2357C38.0952 94.7048 43.5286 100 50 100C56.4714 100 61.9048 94.7048 61.9048 88.2357V61.9048H88.2357C94.7048 61.9048 100 56.4714 100 50C100 43.5286 94.7048 38.0952 88.2357 38.0952Z"
              fill="#4CAF50"
            />
          </svg>
        </div>
        <div className="col col--10">
          <h3>Add your own plugin</h3>
          <p>You can simply add your own plugin to this index.</p>
        </div>
      </div>
    </a>
  );

  const fuse = new Fuse(plugins, {
    keys: ['name', 'description'],
    threshold: 0.3,
  });

  const [query, setQuery] = useState('');

  const searchedPlugins = query.length >= 1 ? fuse.search(query) : plugins.map((p) => ({ item: p }));

  return (
    <Layout title="Woodpecker CI plugins" description="List of all Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div style={{ display: 'flex', flexFlow: 'column', alignItems: 'center' }}>
            <h1>Woodpecker CI plugins</h1>
            <p>This list contains plugins which you can use to easily execute usual pipeline tasks.</p>
            <a href={applyForIndexUrl} target="_blank" rel="noopener noreferrer" className="button button--primary">
              🙏 Please add your plugin
            </a>
          </div>
          <div className="container" style={{ display: 'flex', flexFlow: 'column', marginTop: '4rem' }}>
            <input
              type="search"
              autoComplete="off"
              value={query}
              onChange={(event) => setQuery(event.currentTarget.value)}
              placeholder="Search for a plugin ..."
              className="wp-plugin-search"
            />
            <div className="wp-plugins-list">
              {/* {query.length == 0 && <NewPluginPanel />} */}
              {searchedPlugins.map((plugin, idx) => (
                <PluginPanel key={idx} plugin={plugin.item} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default WoodpeckerPluginList;
