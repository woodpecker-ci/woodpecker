import React, { Component } from "react";

export default class ExpandIcon extends Component {
	render() {
		return (
			<svg
				className={this.props.className}
				width={this.props.size || 24}
				height={this.props.size || 24}
				viewBox="0 0 24 24"
			>
				<path d="M7.41 7.84L12 12.42l4.59-4.58L18 9.25l-6 6-6-6z" />
				<path d="M0-.75h24v24H0z" fill="none" />
			</svg>
		);
	}
}
