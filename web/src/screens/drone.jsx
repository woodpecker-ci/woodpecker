import React, { Component } from "react";

import { root } from "baobab-react/higher-order";
import tree from "~/config/state";
import client from "~/config/client";
import { drone } from "~/config/client/inject";
import { LoginForm, LoginError } from "~/screens/login/screens";
import Title from "./titles";
import Layout from "./layout";
import RedirectRoot from "./redirect";
import { fetchFeedOnce, subscribeToFeedOnce } from "~/shared/utils/feed";

import { BrowserRouter, Route, Switch } from "react-router-dom";

// eslint-disable-next-line no-unused-vars
import styles from "./drone.less";

if (module.hot) {
  require("preact/devtools");
}

class App extends Component {
  render() {
    return (
      <BrowserRouter>
        <div>
          <Title />
          <Switch>
            <Route path="/" exact={true} component={RedirectRoot} />
            <Route path="/login/form" exact={true} component={LoginForm} />
            <Route path="/login/error" exact={true} component={LoginError} />
            <Route path="/" exact={false} component={Layout} />
          </Switch>
        </div>
      </BrowserRouter>
    );
  }
}

if (tree.exists(["user", "data"])) {
  fetchFeedOnce(tree, client);
  subscribeToFeedOnce(tree, client);
}

client.onerror = error => {
  console.error(error);
  if (error.status === 401) {
    tree.unset(["user", "data"]);
  }
};

export default root(tree, drone(client, App));
