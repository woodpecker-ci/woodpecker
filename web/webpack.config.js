require("dotenv").config();

var path = require("path");
var webpack = require("webpack");
var HtmlWebpackPlugin = require("html-webpack-plugin");

const ENV = process.env.NODE_ENV || "development";

module.exports = {
  entry: {
    app: "./src",
    vendor: [
      "ansi_up",
      "babel-polyfill",
      "baobab",
      "baobab-react",
      "classnames",
      "drone-js",
      "humanize-duration",
      "preact",
      "preact-compat",
      "query-string",
      "react-router",
      "react-router-dom",
      "react-screen-size",
      "react-timeago",
      "react-title-component",
      "react-transition-group"
    ]
  },

  // where to dump the output of a production build
  output: {
    publicPath: "/",
    path: path.join(__dirname, "dist/files"),
    filename: "static/bundle.[chunkhash].js"
  },

  resolve: {
    alias: {
      client: path.resolve(__dirname, "src/client/"),
      config: path.resolve(__dirname, "src/config/"),
      components: path.resolve(__dirname, "src/components/"),
      layouts: path.resolve(__dirname, "src/layouts/"),
      pages: path.resolve(__dirname, "src/pages/"),
      screens: path.resolve(__dirname, "src/screens/"),
      shared: path.resolve(__dirname, "src/shared/"),

      react: "preact-compat/dist/preact-compat",
      "react-dom": "preact-compat/dist/preact-compat",
      "create-react-class": "preact-compat/lib/create-react-class"
    }
  },

  module: {
    rules: [
      {
        test: /\.jsx?/i,
        exclude: /node_modules/,
        loader: "babel-loader"
      },

      {
        test: /\.(less|css)$/,
        loader: "style-loader"
      },

      {
        test: /\.(less|css)$/,
        loader: "css-loader",
        query: {
          modules: true,
          localIdentName: "[name]__[local]___[hash:base64:5]"
        }
      },

      {
        test: /\.(less|css)$/,
        loader: "less-loader"
      }
    ]
  },

  plugins: [
    new webpack.optimize.CommonsChunkPlugin({
      name: "vendor",
      filename: "static/vendor.[hash].js"
    }),
    new HtmlWebpackPlugin({
      favicon: "src/public/favicon.svg",
      template: "src/index.html"
    })
  ].concat(
    ENV === "production"
      ? [
          new webpack.optimize.UglifyJsPlugin({
            output: {
              comments: false
            },
            exclude: [/bundle/],
            compress: {
              unsafe_comps: true,
              properties: true,
              keep_fargs: false,
              pure_getters: true,
              collapse_vars: true,
              unsafe: true,
              warnings: false,
              screw_ie8: true,
              sequences: true,
              dead_code: true,
              drop_debugger: true,
              comparisons: true,
              conditionals: true,
              evaluate: true,
              booleans: true,
              loops: true,
              unused: true,
              hoist_funs: true,
              if_return: true,
              join_vars: true,
              cascade: true,
              drop_console: true
            }
          })
        ]
      : []
  ),

  devServer: {
    port: process.env.PORT || 9999,

    // serve up any static files from src/
    contentBase: path.join(__dirname, "src"),

    // enable gzip compression:
    compress: true,

    // enable pushState() routing, as used by preact-router et al:
    historyApiFallback: true
  }
};
