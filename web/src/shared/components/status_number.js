import React, { Component } from "react";
import classnames from "classnames";

import styles from "./status_number.less";

export default class StatusNumber extends Component {
  render() {
    const { status, number } = this.props;
    const className = classnames(styles.root, styles[status]);
    return <div className={className}>{number}</div>;
  }
}
