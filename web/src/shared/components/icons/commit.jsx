import React, { Component } from "react";

export default class CommitIcon extends Component {
  render() {
    return (
      <svg
        className={this.props.className}
        width={this.props.size || 24}
        height={this.props.size || 24}
        viewBox="0 0 24 24"
      >
        <path d="M17,12C17,14.42 15.28,16.44 13,16.9V21H11V16.9C8.72,16.44 7,14.42 7,12C7,9.58 8.72,7.56 11,7.1V3H13V7.1C15.28,7.56 17,9.58 17,12M12,9A3,3 0 0,0 9,12A3,3 0 0,0 12,15A3,3 0 0,0 15,12A3,3 0 0,0 12,9Z" />
      </svg>
    );
  }
}
