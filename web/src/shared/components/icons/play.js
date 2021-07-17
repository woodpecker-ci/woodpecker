import React, { Component } from "react";

export default class PlayIcon extends Component {
  render() {
    return (
      <svg
        className={this.props.className}
        width={this.props.size || 24}
        height={this.props.size || 24}
        viewBox="0 0 24 24"
      >
        <path d="M8 5v14l11-7z" />
        <path d="M0 0h24v24H0z" fill="none" />
      </svg>
    );
  }
}
