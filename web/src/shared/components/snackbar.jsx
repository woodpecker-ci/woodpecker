import React from "react";
import styles from "./snackbar.less";
import CloseIcon from "shared/components/icons/close";
import { CSSTransitionGroup } from "react-transition-group";

export class Snackbar extends React.Component {
  render() {
    const { message } = this.props;

    let classes = [styles.snackbar];
    if (message) {
      classes.push(styles.open);
    }

    const content = message ? (
      <div className={classes.join(" ")} key={message}>
        <div>{message}</div>
        <button onClick={this.props.onClose}>
          <CloseIcon />
        </button>
      </div>
    ) : null;

    return (
      <CSSTransitionGroup
        transitionName="slideup"
        transitionEnterTimeout={200}
        transitionLeaveTimeout={200}
        transitionAppearTimeout={200}
        transitionAppear={true}
        transitionEnter={true}
        transitionLeave={true}
        className={classes.root}
      >
        {content}
      </CSSTransitionGroup>
    );
  }
}

// const SnackbarContent = ({ children, ...props }) => {
// 	<div {...props}>{children}</div>
// }
//
// const SnackbarClose = ({ children, ...props }) => {
// 	<div {...props}>{children}</div>
// }
