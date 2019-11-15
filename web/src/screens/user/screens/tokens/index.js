import React, { Component } from "react";

import { generateToken } from "shared/utils/users";
import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";
import styles from "./index.less";

const binding = (props, context) => {
  return {
    location: ["location"],
    token: ["token"],
  };
};

@inject
@branch(binding)
export default class Tokens extends Component {
  shouldComponentUpdate(nextProps, nextState) {
    return (
      this.props.location !== nextProps.location ||
      this.props.token !== nextProps.token
    );
  }

  componentWillMount() {
    const { drone, dispatch } = this.props;

    dispatch(generateToken, drone);
  }

  render() {
    const { location, token } = this.props;

    if (!location || !token) {
      return <div>Loading</div>;
    }
    return (
      <div className={styles.root}>
        <h2>Your Personal Token:</h2>
        <pre>{token}</pre>
        <h2>Example API Usage:</h2>
        <pre>{usageWithCURL(location, token)}</pre>
        <h2>Example CLI Usage:</h2>
        <pre>{usageWithCLI(location, token)}</pre>
      </div>
    );
  }
}

const usageWithCURL = (location, token) => {
  return `curl -i ${location.protocol}//${location.host}/api/user -H "Authorization: Bearer ${token}"`;
};

const usageWithCLI = (location, token) => {
  return `export DRONE_SERVER=${location.protocol}//${location.host}
		export DRONE_TOKEN=${token}

		drone info`;
};
