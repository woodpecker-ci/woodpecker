import React, { Component } from "react";
import CloseIcon from "shared/components/icons/close";
import styles from "./drawer.less";
import { CSSTransitionGroup } from "react-transition-group";

export const DOCK_LEFT = styles.left;
export const DOCK_RIGHT = styles.right;

export class Drawer extends Component {
  render() {
    const { open, position } = this.props;

    let classes = [styles.drawer];
    if (open) {
      classes.push(styles.open);
    }
    if (position) {
      classes.push(position);
    }

    var child = open ? (
      <div key={0} onClick={this.props.onClick} className={styles.backdrop} />
    ) : null;

    return (
      <div className={classes.join(" ")}>
        <CSSTransitionGroup
          transitionName="fade"
          transitionEnterTimeout={150}
          transitionLeaveTimeout={150}
          transitionAppearTimeout={150}
          transitionAppear={true}
          transitionEnter={true}
          transitionLeave={true}
        >
          {child}
        </CSSTransitionGroup>
        <div className={styles.inner}>{this.props.children}</div>
      </div>
    );
  }
}

export class CloseButton extends Component {
  render() {
    return (
      <button className={styles.close} onClick={this.props.onClick}>
        <CloseIcon />
      </button>
    );
  }
}

export class MenuButton extends Component {
  render() {
    return (
      <button className={styles.close} onClick={this.props.onClick}>
        Show Menu
      </button>
    );
  }
}
