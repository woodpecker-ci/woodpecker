import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import { WoodpeckerPlugin as WoodpeckerPluginType } from '../types';

export function WoodpeckerPlugin({ plugin }: { plugin: WoodpeckerPluginType }) {
  return (
    <Layout
      title="Woodpecker CI plugins"
      description="List of Woodpecker-CI plugins"
    >
      <main className={clsx("container margin-vert--lg")}>
        <section>
          <div className={clsx("container")}>
            <a href="/plugins">&lt;&lt; Back to plugin list</a>
            <div className={clsx("row")}>
              <div className={clsx("col col--10")}>
                <h1>{plugin.name}</h1>
                <p>{plugin.description}</p>
                <a href={plugin.url} target="_blank" rel="noopener noreferrer">
                  {plugin.url}
                </a>
              </div>
              <div className={clsx("col col--2")}>
                <img src={plugin.icon} width="150" height="150" />
              </div>
            </div>
            <hr />
            <div dangerouslySetInnerHTML={{ __html: plugin.docs }} />
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default WoodpeckerPlugin;
