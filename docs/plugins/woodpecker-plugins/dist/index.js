"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
exports.__esModule = true;
exports.getSwizzleComponentList = void 0;
var rest_1 = require("@octokit/rest");
var path_1 = __importDefault(require("path"));
var markdown = __importStar(require("./markdown"));
function loadContent() {
    return __awaiter(this, void 0, void 0, function () {
        var octokit, codeResults, plugins;
        var _this = this;
        return __generator(this, function (_a) {
            switch (_a.label) {
                case 0:
                    octokit = new rest_1.Octokit();
                    return [4 /*yield*/, octokit.rest.search.code({
                            // search for repos in woodpecker-ci org with a file docs.md containing the string WOODPECKER_PLUGIN_DOCS
                            q: 'org:woodpecker-ci filename:docs.md WOODPECKER_PLUGIN_DOCS'
                        })];
                case 1:
                    codeResults = (_a.sent()).data.items;
                    return [4 /*yield*/, Promise.all(codeResults
                            .filter(function (i) { return i.repository.name.startsWith('plugin-'); })
                            .map(function (i) { return __awaiter(_this, void 0, void 0, function () {
                            var docsResult, docs, header, plugin;
                            return __generator(this, function (_a) {
                                switch (_a.label) {
                                    case 0: return [4 /*yield*/, octokit.repos.getContent({
                                            owner: 'woodpecker-ci',
                                            repo: i.repository.name,
                                            path: '/docs.md'
                                        })];
                                    case 1:
                                        docsResult = (_a.sent()).data;
                                        docs = Buffer.from(docsResult.content, 'base64').toString('ascii');
                                        header = markdown.getHeader(docs);
                                        plugin = {
                                            name: (header === null || header === void 0 ? void 0 : header.name) || i.repository.name,
                                            repoName: i.repository.name,
                                            url: i.repository.html_url,
                                            icon: header === null || header === void 0 ? void 0 : header.icon,
                                            description: header === null || header === void 0 ? void 0 : header.description,
                                            docs: markdown.getContent(docs)
                                        };
                                        return [2 /*return*/, plugin];
                                }
                            });
                        }); }))];
                case 2:
                    plugins = _a.sent();
                    return [2 /*return*/, {
                            plugins: plugins
                        }];
            }
        });
    });
}
function contentLoaded(_a) {
    var plugins = _a.content.plugins, actions = _a.actions;
    return __awaiter(this, void 0, void 0, function () {
        var createData, addRoute, pluginsJsonPath;
        var _this = this;
        return __generator(this, function (_b) {
            switch (_b.label) {
                case 0:
                    createData = actions.createData, addRoute = actions.addRoute;
                    return [4 /*yield*/, createData('plugins.json', JSON.stringify(plugins))];
                case 1:
                    pluginsJsonPath = _b.sent();
                    return [4 /*yield*/, Promise.all(plugins.map(function (plugin) { return __awaiter(_this, void 0, void 0, function () {
                            var pluginJsonPath;
                            return __generator(this, function (_a) {
                                switch (_a.label) {
                                    case 0: return [4 /*yield*/, createData("plugin-" + plugin.repoName + ".json", JSON.stringify(plugin))];
                                    case 1:
                                        pluginJsonPath = _a.sent();
                                        addRoute({
                                            path: "/plugins/" + plugin.repoName,
                                            component: '@theme/WoodpeckerPlugin',
                                            modules: {
                                                plugin: pluginJsonPath
                                            },
                                            exact: true
                                        });
                                        return [2 /*return*/];
                                }
                            });
                        }); }))];
                case 2:
                    _b.sent();
                    addRoute({
                        path: '/plugins',
                        component: '@theme/WoodpeckerPluginList',
                        modules: {
                            plugins: pluginsJsonPath
                        },
                        exact: true
                    });
                    return [2 /*return*/];
            }
        });
    });
}
function pluginWoodpeckerPluginsIndex(context, options) {
    return {
        name: 'woodpecker-plugins',
        loadContent: loadContent,
        contentLoaded: contentLoaded,
        getThemePath: function () {
            return path_1["default"].join(__dirname, '..', 'dist', 'theme');
        },
        getTypeScriptThemePath: function () {
            return path_1["default"].join(__dirname, '..', 'src', 'theme');
        },
        getPathsToWatch: function () {
            return [path_1["default"].join(__dirname, '..', 'dist', '**', '*.{js,jsx}')];
        }
    };
}
exports["default"] = pluginWoodpeckerPluginsIndex;
var getSwizzleComponentList = function () {
    return ['WoodpeckerPluginList', 'WoodpeckerPlugin'];
};
exports.getSwizzleComponentList = getSwizzleComponentList;
