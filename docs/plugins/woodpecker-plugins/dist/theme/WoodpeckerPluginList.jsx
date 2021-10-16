"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
exports.__esModule = true;
exports.WoodpeckerPluginList = void 0;
var react_1 = __importDefault(require("react"));
var clsx_1 = __importDefault(require("clsx"));
var Layout_1 = __importDefault(require("@theme/Layout"));
function PluginPanel(_a) {
    var plugin = _a.plugin;
    var pluginUrl = "/plugins/" + plugin.repoName;
    return (<div className={(0, clsx_1["default"])("col col--6")}>
      <div className={(0, clsx_1["default"])("card margin-horiz--sm margin-vert--md ")}>
        <div className={(0, clsx_1["default"])("card__header row")}>
          <div className={(0, clsx_1["default"])("col col--8")}>
            <a href={pluginUrl}>
              <h3>{plugin.name}</h3>
            </a>
            <p>{plugin.description}</p>
          </div>
          <div className={(0, clsx_1["default"])("col col--4 text--right")}>
            <img src={plugin.icon} width="100" height="100"/>
          </div>
        </div>
        <div className={(0, clsx_1["default"])("card__footer")}>
          <a href={pluginUrl} className={(0, clsx_1["default"])('button button--secondary button--outline button--block ')}>
            Open {plugin.name}
          </a>
        </div>
      </div>
    </div>);
}
function WoodpeckerPluginList(_a) {
    var plugins = _a.plugins;
    return (<Layout_1.default title="Woodpecker CI plugins" description="List of all Woodpecker-CI plugins">
      <main className="container margin-vert--lg">
        <section>
          <div className="container">
            <div className="row">
              {plugins.map(function (plugin, idx) { return (<PluginPanel key={idx} plugin={plugin}/>); })}
            </div>
          </div>
        </section>
      </main>
    </Layout_1.default>);
}
exports.WoodpeckerPluginList = WoodpeckerPluginList;
// FIX to resolve "React is not defined"
window.React = react_1["default"];
exports["default"] = WoodpeckerPluginList;
