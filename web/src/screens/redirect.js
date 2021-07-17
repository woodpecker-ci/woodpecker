import React, { Component } from "react";
import { Redirect } from "react-router-dom";
import { branch } from "baobab-react/higher-order";
import { Message } from "shared/components/sync";

const binding = (props, context) => {
  return {
    feed: ["feed"],
    user: ["user", "data"],
    syncing: ["user", "syncing"],
  };
};

@branch(binding)
export default class RedirectRoot extends Component {
  componentWillReceiveProps(nextProps) {
    const { user } = nextProps;
    if (!user && window) {
      window.location.href = "/login?url=" + window.location.href;
    }
  }

  render() {
    const { user, syncing } = this.props;
    const { latest, loaded } = this.props.feed;

    return !loaded && syncing ? (
      <Message />
    ) : !loaded ? (
      undefined
    ) : !user ? (
      undefined
    ) : !latest ? (
      <Redirect to="/account/repos" />
    ) : !latest.number ? (
      <Redirect to={`/${latest.full_name}`} />
    ) : (
      <Redirect to={`/${latest.full_name}/${latest.number}`} />
    );
  }
}
