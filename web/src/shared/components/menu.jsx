import React, { Component } from "react";
import { NavLink as Link } from "react-router-dom";
import PropTypes from "prop-types";

import styles from "./menu.less";

export default class Menu extends Component {
  propTypes = { items: PropTypes.array, right: PropTypes.any };
  render() {
    const items = this.props.items;
    const right = this.props.right ? (
      <div className={styles.right}>{this.props.right}</div>
    ) : null;
    return (
      <section className={styles.root}>
        <div className={styles.left}>
          {items.map(i => (
            <Link
              key={i.to + i.label}
              to={i.to}
              exact={true}
              activeClassName={styles["link-active"]}
            >
              {i.label}
            </Link>
          ))}
        </div>
        {right}
      </section>
    );
  }
}
