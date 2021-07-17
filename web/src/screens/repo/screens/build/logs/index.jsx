import React, { Component } from "react";
import { inject } from "config/client/inject";
import { branch } from "baobab-react/higher-order";
import { repositorySlug } from "shared/utils/repository";
import { assertProcFinished, assertProcRunning } from "shared/utils/proc";
import { fetchLogs, subscribeToLogs, toggleLogs } from "shared/utils/logs";

import Term from "./components/term";

import { Top, Bottom, scrollToTop, scrollToBottom } from "./components/anchor";

import { ExpandIcon, PauseIcon, PlayIcon } from "shared/components/icons/index";

import styles from "./index.less";

const binding = (props, context) => {
  const { owner, repo, build } = props.match.params;
  const slug = repositorySlug(owner, repo);
  const number = parseInt(build);
  const pid = parseInt(props.proc.pid);

  return {
    logs: ["logs", "data", slug, number, pid, "data"],
    eof: ["logs", "data", slug, number, pid, "eof"],
    loading: ["logs", "data", slug, number, pid, "loading"],
    error: ["logs", "data", slug, number, pid, "error"],
    follow: ["logs", "follow"],
  };
};

@inject
@branch(binding)
export default class Output extends Component {
  constructor(props, context) {
    super(props, context);
    this.handleFollow = this.handleFollow.bind(this);
  }

  componentWillMount() {
    if (this.props.proc) {
      this.componentWillUpdate(this.props);
    }
  }

  componentWillUpdate(nextProps) {
    const { loading, logs, eof, error } = nextProps;
    const routeChange = this.props.match.url !== nextProps.match.url;

    if (loading || error || (logs && eof)) {
      return;
    }

    if (assertProcFinished(nextProps.proc)) {
      return this.props.dispatch(
        fetchLogs,
        nextProps.drone,
        nextProps.match.params.owner,
        nextProps.match.params.repo,
        nextProps.build.number,
        nextProps.proc.pid,
      );
    }

    if (assertProcRunning(nextProps.proc) && (!logs || routeChange)) {
      this.props.dispatch(
        subscribeToLogs,
        nextProps.drone,
        nextProps.match.params.owner,
        nextProps.match.params.repo,
        nextProps.build.number,
        nextProps.proc,
      );
    }
  }

  componentDidUpdate() {
    if (this.props.follow) {
      scrollToBottom();
    }
  }

  handleFollow() {
    this.props.dispatch(toggleLogs, !this.props.follow);
  }

  render() {
    const { logs, error, proc, loading, follow, highlighted } = this.props;

    if (loading || !proc) {
      return <Term.Loading />;
    }

    if (error) {
      return <Term.Error />;
    }

    return (
      <div>
        <Top />
        <Term
          lines={logs || []}
          highlighted={highlighted}
          exitcode={assertProcFinished(proc) ? proc.exit_code : undefined}
        />
        <Bottom />
        <Actions
          running={assertProcRunning(proc)}
          following={follow}
          onfollow={this.handleFollow}
          onunfollow={this.handleFollow}
        />
      </div>
    );
  }
}

/**
 * Component renders floating log actions. These can be used
 * to follow, unfollow, scroll to top and scroll to bottom.
 */
const Actions = ({ following, running, onfollow, onunfollow }) => (
  <div className={styles.actions}>
    {running && !following ? (
      <button onClick={onfollow} className={styles.follow}>
        <PlayIcon />
      </button>
    ) : null}

    {running && following ? (
      <button onClick={onunfollow} className={styles.unfollow}>
        <PauseIcon />
      </button>
    ) : null}

    <button onClick={scrollToTop} className={styles.bottom}>
      <ExpandIcon />
    </button>

    <button onClick={scrollToBottom} className={styles.top}>
      <ExpandIcon />
    </button>
  </div>
);
