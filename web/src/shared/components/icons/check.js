import React, { Component } from "react";

export default class CheckIcon extends Component {
  render() {
    return (
      <svg
        className={this.props.className}
        width={this.props.size || 24}
        height={this.props.size || 24}
        viewBox="0 0 24 24"
      >
        <path d="M0 0h24v24H0z" fill="none" />
        <path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4L9 16.2z" />
      </svg>
    );
  }
}
