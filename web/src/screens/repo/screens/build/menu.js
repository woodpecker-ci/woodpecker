import React, { Component } from "react";
import RepoMenu from "../builds/menu";
import { RefreshIcon, CloseIcon } from "shared/components/icons";

import { cancelBuild, restartBuild } from "shared/utils/build";
import { findChildProcess } from "shared/utils/proc";
import { repositorySlug } from "shared/utils/repository";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

const binding = (props, context) => {
  const { owner, repo, build } = props.match.params;
  const slug = repositorySlug(owner, repo);
  const number = parseInt(build);
  return {
    repo: ["repos", "data", slug],
    build: ["builds", "data", slug, number]
  };
};

@inject
@branch(binding)
export default class BuildMenu extends Component {
  constructor(props, context) {
    super(props, context);

    this.handleCancel = this.handleCancel.bind(this);
    this.handleRestart = this.handleRestart.bind(this);
  }

  handleRestart() {
    const { dispatch, drone, repo, build } = this.props;
    dispatch(restartBuild, drone, repo.owner, repo.name, build.number);
  }

  handleCancel() {
    const { dispatch, drone, repo, build, match } = this.props;
    const proc = findChildProcess(build.procs, match.params.proc || 2);

    dispatch(
      cancelBuild,
      drone,
      repo.owner,
      repo.name,
      build.number,
      proc.ppid
    );
  }

  render() {
    const { build } = this.props;

    const rightSide = !build ? (
      undefined
    ) : (
      <section>
        {build.status === "pending" || build.status === "running" ? (
          <button onClick={this.handleCancel}>
            <CloseIcon />
            <span>Cancel</span>
          </button>
        ) : (
          <button onClick={this.handleRestart}>
            <RefreshIcon />
            <span>Restart Build</span>
          </button>
        )}
      </section>
    );

    return (
      <div>
        <RepoMenu {...this.props} right={rightSide} />
      </div>
    );
  }
}
