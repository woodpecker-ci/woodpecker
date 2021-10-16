import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
function PluginPanel(_a) {
    var plugin = _a.plugin;
    var pluginUrl = "/plugins/" + plugin.repoName;
    return (<div className={clsx("col col--6")}>
      <div className={clsx("card margin-horiz--sm margin-vert--md ")}>
        <div className={clsx("card__header row")}>
          <div className={clsx("col col--8")}>
            <a href={pluginUrl}>
              <h3>{plugin.name}</h3>
            </a>
            <p>{plugin.description}</p>
          </div>
          <div className={clsx("col col--4 text--right")}>
            <img src={plugin.icon} width="100" height="100"/>
          </div>
        </div>
        <div className={clsx("card__footer")}>
          <a href={pluginUrl} className={clsx('button button--secondary button--outline button--block ')}>
            Open {plugin.name}
          </a>
        </div>
      </div>
    </div>);
}
export function WoodpeckerPluginList(_a) {
    var plugins = _a.plugins;
    return (<Layout title="Woodpecker CI plugins" description="List of all Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div className="container">
            <div className="row">
              {plugins.map(function (plugin, idx) { return (<PluginPanel key={idx} plugin={plugin}/>); })}
            </div>
          </div>
        </section>
      </main>
    </Layout>);
}
// if (typeof window !== 'undefined' && React) {
//   // FIX to resolve "React is not defined"
//   window.React = React;
// }
export default WoodpeckerPluginList;
