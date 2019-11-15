import React, { Component } from "react";
import styles from "./switch.less";

export class Switch extends Component {
  render() {
    const { checked, onchange } = this.props;
    return (
      <label className={styles.switch}>
        <input type="checkbox" checked={checked} onChange={onchange} />
      </label>
    );
  }
}
