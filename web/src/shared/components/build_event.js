import React, { Component } from "react";
import {
  BranchIcon,
  CommitIcon,
  DeployIcon,
  LaunchIcon,
  MergeIcon,
  TagIcon,
} from "shared/components/icons/index";
import {
  EVENT_TAG,
  EVENT_PULL_REQUEST,
  EVENT_DEPLOY,
} from "shared/constants/events";

import styles from "./build_event.less";

export default class BuildEvent extends Component {
  render() {
    const { event, branch, commit, refs, refspec, link, target } = this.props;

    return (
      <div className={styles.host}>
        <div className={styles.row}>
          <div>
            <CommitIcon />
          </div>
          <div>{commit && commit.substr(0, 10)}</div>
        </div>
        <div className={styles.row}>
          <div>
            {event === EVENT_TAG ? (
              <TagIcon />
            ) : event === EVENT_PULL_REQUEST ? (
              <MergeIcon />
            ) : event === EVENT_DEPLOY ? (
              <DeployIcon />
            ) : (
              <BranchIcon />
            )}
          </div>
          <div>
            {event === EVENT_TAG && refs ? (
              trimTagRef(refs)
            ) : event === EVENT_PULL_REQUEST && refspec ? (
              trimMergeRef(refs)
            ) : event === EVENT_DEPLOY && target ? (
              target
            ) : (
              branch
            )}
          </div>
        </div>
        <a href={link} target="_blank">
          <LaunchIcon />
        </a>
      </div>
    );
  }
}

const trimMergeRef = ref => {
  return ref.match(/\d/g) || ref;
};

const trimTagRef = ref => {
  return ref.startsWith("refs/tags/") ? ref.substr(10) : ref;
};

// push
// pull request (ref)
// tag (ref)
// deploy
