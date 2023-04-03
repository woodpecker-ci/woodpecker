import React, { useState, useRef } from 'react';
import Fuse from 'fuse.js';
import Layout from '@theme/Layout';
import './style.css';
import { WoodpeckerPlugin } from '../types';
import { IconPlugin, IconVerified } from './Icons';

function PluginPanel({ plugin }: { plugin: WoodpeckerPlugin }) {
  const pluginUrl = `/plugins/${plugin.name}`;

  return (
    <a href={pluginUrl} className="card shadow--md wp-plugin-card">
      <div className="card__header row">
        <div className="col col--2 text--left">{plugin.icon ? <img src={plugin.icon} width="50" /> : IconPlugin()}</div>
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

  const [query, setQuery] = useState('');

  const fuse = useRef(new Fuse(plugins, {
    keys: ['name', 'description'],
    threshold: 0.3,
  }));

  const searchedPlugins = query.length >= 1 ? fuse.current.search(query).map((p) => ( p.item )) : plugins;

  return (
    <Layout title="Woodpecker CI plugins" description="List of all Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div style={{ display: 'flex', flexFlow: 'column', alignItems: 'center' }}>
            <h1>Woodpecker CI plugins</h1>
            <p>This list contains plugins which you can use to easily execute usual pipeline tasks.</p>
            <a href={applyForIndexUrl} target="_blank" rel="noopener noreferrer" className="button button--primary">
              🎉 Add your plugin
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
              {searchedPlugins.map((plugin) => (
                <PluginPanel key={plugin.name} plugin={plugin} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default WoodpeckerPluginList;
