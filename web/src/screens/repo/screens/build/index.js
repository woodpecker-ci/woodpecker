import React, { Component } from "react";
import { Link } from "react-router-dom";

import { fetchBuild, approveBuild, declineBuild } from "shared/utils/build";
import {
  STATUS_BLOCKED,
  STATUS_DECLINED,
  STATUS_ERROR,
} from "shared/constants/status";

import { findChildProcess } from "shared/utils/proc";
import { fetchRepository } from "shared/utils/repository";

import Breadcrumb, { SEPARATOR } from "shared/components/breadcrumb";

import { Approval, Details, ProcList } from "./components";

import { branch } from "baobab-react/higher-order";
import { inject } from "config/client/inject";

import Output from "./logs";

import styles from "./index.less";

const binding = (props, context) => {
  const { owner, repo, build } = props.match.params;
  const slug = `${owner}/${repo}`;
  const number = parseInt(build);

  return {
    repo: ["repos", "data", slug],
    build: ["builds", "data", slug, number],
  };
};

@inject
@branch(binding)
export default class BuildLogs extends Component {
  constructor(props, context) {
    super(props, context);

    this.handleApprove = this.handleApprove.bind(this);
    this.handleDecline = this.handleDecline.bind(this);
  }

  componentWillMount() {
    this.synchronize(this.props);
  }

  handleApprove() {
    const { repo, build, drone } = this.props;
    this.props.dispatch(
      approveBuild,
      drone,
      repo.owner,
      repo.name,
      build.number,
    );
  }

  handleDecline() {
    const { repo, build, drone } = this.props;
    this.props.dispatch(
      declineBuild,
      drone,
      repo.owner,
      repo.name,
      build.number,
    );
  }

  componentWillUpdate(nextProps) {
    if (this.props.match.url !== nextProps.match.url) {
      this.synchronize(nextProps);
    }
  }

  synchronize(props) {
    if (!props.repo) {
      this.props.dispatch(
        fetchRepository,
        props.drone,
        props.match.params.owner,
        props.match.params.repo,
      );
    }
    if (!props.build || !props.build.procs) {
      this.props.dispatch(
        fetchBuild,
        props.drone,
        props.match.params.owner,
        props.match.params.repo,
        props.match.params.build,
      );
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    return this.props !== nextProps;
  }

  render() {
    const { repo, build } = this.props;

    if (!build || !repo) {
      return this.renderLoading();
    }

    if (build.status === STATUS_DECLINED || build.status === STATUS_ERROR) {
      return this.renderError();
    }

    if (build.status === STATUS_BLOCKED) {
      return this.renderBlocked();
    }

    if (!build.procs) {
      return this.renderLoading();
    }

    return this.renderSimple();
  }

  renderLoading() {
    return (
      <div className={styles.host}>
        <div className={styles.columns}>
          <div className={styles.right}>Loading ...</div>
        </div>
      </div>
    );
  }

  renderBlocked() {
    const { build } = this.props;
    return (
      <div className={styles.host}>
        <div className={styles.columns}>
          <div className={styles.right}>
            <Details build={build} />
          </div>
          <div className={styles.left}>
            <Approval
              onapprove={this.handleApprove}
              ondecline={this.handleDecline}
            />
          </div>
        </div>
      </div>
    );
  }

  renderError() {
    const { build } = this.props;
    return (
      <div className={styles.host}>
        <div className={styles.columns}>
          <div className={styles.right}>
            <Details build={build} />
          </div>
          <div className={styles.left}>
            <div className={styles.logerror}>
              {build.status === STATUS_ERROR ? (
                build.error
              ) : (
                "Pipeline execution was declined"
              )}
            </div>
          </div>
        </div>
      </div>
    );
  }

  highlightedLine() {
    if (location.hash.startsWith("#L")) {
      return parseInt(location.hash.substr(2)) - 1;
    }

    return undefined;
  }

  renderSimple() {
    // if (nextProps.build.procs[0].children !== undefined){
    // 	return null;
    // }

    const { repo, build, match } = this.props;
    const selectedProc = match.params.proc
      ? findChildProcess(build.procs, match.params.proc)
      : build.procs[0].children[0];
    const selectedProcParent = findChildProcess(build.procs, selectedProc.ppid);
    const highlighted = this.highlightedLine();

    return (
      <div className={styles.host}>
        <div className={styles.columns}>
          <div className={styles.right}>
            <Details build={build} />
            <section className={styles.sticky}>
              {build.procs.map(function(rootProc) {
                return (
                  <div style="padding-bottom: 20px;" key={rootProc.pid}>
                    <ProcList
                      key={rootProc.pid}
                      repo={repo}
                      build={build}
                      rootProc={rootProc}
                      selectedProc={selectedProc}
                      renderName={build.procs.length > 1}
                    />
                  </div>
                );
              })}
            </section>
          </div>
          <div className={styles.left}>
            {selectedProc && selectedProc.error ? (
              <div className={styles.logerror}>{selectedProc.error}</div>
            ) : null}
            {selectedProcParent && selectedProcParent.error ? (
              <div className={styles.logerror}>{selectedProcParent.error}</div>
            ) : null}
            <Output
              match={this.props.match}
              build={this.props.build}
              proc={selectedProc}
              highlighted={highlighted}
            />
          </div>
        </div>
      </div>
    );
  }
}

export class BuildLogsTitle extends Component {
  render() {
    const { owner, repo, build } = this.props.match.params;
    return (
      <Breadcrumb
        elements={[
          <Link to={`/${owner}/${repo}`} key={`${owner}-${repo}`}>
            {owner} / {repo}
          </Link>,
          SEPARATOR,
          <Link
            to={`/${owner}/${repo}/${build}`}
            key={`${owner}-${repo}-${build}`}
          >
            {build}
          </Link>,
        ]}
      />
    );
  }
}
