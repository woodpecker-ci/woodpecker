import React, { Component } from "react";

import Status from "shared/components/status";
import BuildTime from "shared/components/build_time";

import styles from "./list.less";

import { StarIcon } from "shared/components/icons/index";

export const List = ({ children }) => (
  <div className={styles.list}>{children}</div>
);

export class Item extends Component {
  constructor(props) {
    super(props);

    this.handleFave = this.handleFave.bind(this);
  }

  handleFave(e) {
    e.preventDefault();
    this.props.onFave(this.props.item.full_name);
  }

  render() {
    const { item, faved } = this.props;
    return (
      <div className={styles.item}>
        <div onClick={this.handleFave}>
          <StarIcon filled={faved} size={16} className={styles.star} />
        </div>
        <div className={styles.header}>
          <div className={styles.title}>{item.full_name}</div>
          <div className={styles.icon}>
            {item.status ? <Status status={item.status} /> : <noscript />}
          </div>
        </div>

        <div className={styles.body}>
          <BuildTime
            start={item.started_at || item.created_at}
            finish={item.finished_at}
          />
        </div>
      </div>
    );
  }

  shouldComponentUpdate(nextProps, nextState) {
    return (
      this.props.item !== nextProps.item || this.props.faved !== nextProps.faved
    );
  }
}
