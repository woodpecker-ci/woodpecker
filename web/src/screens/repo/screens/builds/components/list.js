import React, { Component } from "react";

import Status from "shared/components/status";
import StatusNumber from "shared/components/status_number";
import BuildTime from "shared/components/build_time";
import BuildMeta from "shared/components/build_event";

import styles from "./list.less";

export const List = ({ children }) => (
  <div className={styles.list}>{children}</div>
);

export class Item extends Component {
  render() {
    const { build } = this.props;
    return (
      <div className={styles.item}>
        <div className={styles.icon}>
          <img src={build.author_avatar} />
        </div>

        <div className={styles.body}>
          <h3>{build.message.split("\n")[0]}</h3>
        </div>

        <div className={styles.meta}>
          <BuildMeta
            link={build.link_url}
            event={build.event}
            commit={build.commit}
            branch={build.branch}
            target={build.deploy_to}
            refspec={build.refspec}
            refs={build.ref}
          />
        </div>

        <div className={styles.break} />

        <div className={styles.time}>
          <BuildTime
            start={build.started_at || build.created_at}
            finish={build.finished_at}
          />
        </div>

        <div className={styles.status}>
          <StatusNumber status={build.status} number={build.number} />
          <Status status={build.status} />
        </div>
      </div>
    );
  }
}
