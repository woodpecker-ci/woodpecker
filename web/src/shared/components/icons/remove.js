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
        <path d="M19 13H5v-2h14v2z" />
        <path d="M0 0h24v24H0z" fill="none" />
      </svg>
    );
  }
}
