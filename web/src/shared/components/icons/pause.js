import React, { Component } from "react";

export default class PauseIcon extends Component {
  render() {
    return (
      <svg
        className={this.props.className}
        width={this.props.size || 24}
        height={this.props.size || 24}
        viewBox="0 0 24 24"
      >
        <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
        <path d="M0 0h24v24H0z" fill="none" />
      </svg>
    );
  }
}
