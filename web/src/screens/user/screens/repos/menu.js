import React, { Component } from "react";
import { syncRepostoryList } from "shared/utils/repository";
import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";
import { SyncIcon } from "shared/components/icons";
import Menu from "shared/components/menu";

const binding = (props, context) => {
  return {
    repos: ["repos"]
  };
};

@inject
@branch(binding)
export default class UserReposMenu extends Component {
  constructor(props, context) {
    super(props, context);

    this.handleClick = this.handleClick.bind(this);
  }

  handleClick() {
    const { dispatch, drone } = this.props;
    dispatch(syncRepostoryList, drone);
  }

  render() {
    const { loaded } = this.props.repos;
    const right = (
      <section>
        <button disabled={!loaded} onClick={this.handleClick}>
          <SyncIcon />
          <span>Synchronize</span>
        </button>
      </section>
    );

    return <Menu items={[]} right={right} />;
  }
}
