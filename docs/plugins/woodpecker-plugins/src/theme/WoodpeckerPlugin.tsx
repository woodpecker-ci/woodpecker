import React from 'react';
import Layout from '@theme/Layout';
import { WoodpeckerPlugin as WoodpeckerPluginType } from '../types';
import { IconContainer, IconVerified, IconWebsite } from './Icons';

export function WoodpeckerPlugin({ plugin }: { plugin: WoodpeckerPluginType }) {
  return (
    <Layout title="Woodpecker CI plugins" description="List of Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div className="container">
            <div className="wp-plugin-breadcrumbs">
              <a href="/plugins">Plugins</a>
              <span> / </span>
              <span>{plugin.name}</span>
            </div>
            <div className="row">
              <div className="col col--10">
                <div style={{ display: 'flex', alignItems: 'center' }}>
                  <h1 style={{ marginBottom: 0 }}>{plugin.name}</h1>
                  {plugin.verified && IconVerified()}
                </div>
                {plugin.author && <span>by {plugin.author}</span>}

                <div style={{ marginTop: '1rem' }}>
                  {plugin.containerImage && (
                    <div style={{ display: 'flex', gap: '.5rem', alignItems: 'center' }}>
                      {IconContainer(20)}
                      {plugin.containerImageUrl ? (
                        <a href={plugin.containerImageUrl} target="_blank" rel="noopener noreferrer">
                          {plugin.containerImage}
                        </a>
                      ) : (
                        <span>{plugin.containerImage}</span>
                      )}
                    </div>
                  )}

                  {plugin.url && (
                    <a
                      href={plugin.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ display: 'flex', gap: '.5rem', alignItems: 'center' }}
                    >
                      <div style={{ color: 'var(--ifm-font-color-base)' }}>{IconWebsite(20)}</div> Website
                    </a>
                  )}

                  {plugin.tags && (
                    <div className="wp-plugin-tags" style={{ marginTop: '.5rem' }}>
                      {plugin.tags.map((tag, idx) => (
                        <span className="badge badge--success" key={idx}>
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                <p style={{ marginTop: '2rem', marginBottom: '1rem' }}>{plugin.description}</p>
              </div>
              { plugin.icon ? <div className="col col--2"> <img src={plugin.icon} width="150" height="150"/> </div> : '' }
            </div>
            <hr style={{ margin: '1rem 0' }} />
            <div dangerouslySetInnerHTML={{ __html: plugin.docs }} />
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default WoodpeckerPlugin;
