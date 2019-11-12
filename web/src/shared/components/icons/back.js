import React, { Component } from "react";

export default class BackIcon extends Component {
	render() {
		return (
			<svg
				className={this.props.className}
				width={this.props.size || 24}
				height={this.props.size || 24}
				viewBox="0 0 24 24"
			>
				<path d="M0 0h24v24H0z" fill="none" />
				<path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z" />
			</svg>
		);
	}
}
