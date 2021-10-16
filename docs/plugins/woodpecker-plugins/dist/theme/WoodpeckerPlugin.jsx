"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
exports.__esModule = true;
exports.WoodpeckerPlugin = void 0;
var react_1 = __importDefault(require("react"));
var clsx_1 = __importDefault(require("clsx"));
var Layout_1 = __importDefault(require("@theme/Layout"));
function WoodpeckerPlugin(_a) {
    var plugin = _a.plugin;
    return (<Layout_1.default title="Woodpecker CI plugins" description="List of Woodpecker-CI plugins">
      <main className={(0, clsx_1["default"])("container margin-vert--lg")}>
        <section>
          <div className={(0, clsx_1["default"])("container")}>
            <a href="/plugins">&lt;&lt; Back to plugin list</a>
            <div className={(0, clsx_1["default"])("row")}>
              <div className={(0, clsx_1["default"])("col col--10")}>
                <h1>{plugin.name}</h1>
                <p>{plugin.description}</p>
                <a target="_blank" rel="noopener noreferrer" href={plugin.url}>
                  {plugin.url}
                </a>
              </div>
              <div className={(0, clsx_1["default"])("col col--2")}>
                <img src={plugin.icon} width="150" height="150"/>
              </div>
            </div>
            <hr />
            <div dangerouslySetInnerHTML={{ __html: plugin.docs }}/>
          </div>
        </section>
      </main>
    </Layout_1.default>);
}
exports.WoodpeckerPlugin = WoodpeckerPlugin;
// FIX to resolve "React is not defined"
window.React = react_1["default"];
exports["default"] = WoodpeckerPlugin;
