import React, { Component } from "react";
import { ScheduleIcon, TimelapseIcon } from "shared/components/icons/index";

import TimeAgo from "react-timeago";
import Duration from "./duration";

import styles from "./build_time.less";

export default class Runtime extends Component {
  render() {
    const { start, finish } = this.props;
    return (
      <div className={styles.host}>
        <div className={styles.row}>
          <div>
            <ScheduleIcon />
          </div>
          <div>{start ? <TimeAgo date={start * 1000} /> : <span>--</span>}</div>
        </div>
        <div className={styles.row}>
          <div>
            <TimelapseIcon />
          </div>
          <div>
            {finish ? (
              <Duration start={start} finished={finish} />
            ) : start ? (
              <TimeAgo date={start * 1000} />
            ) : (
              <span>--</span>
            )}
          </div>
        </div>
      </div>
    );
  }
}
