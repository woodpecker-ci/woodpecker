import React, { Component } from "react";

export default class MenuIcon extends Component {
	render() {
		return (
			<svg
				className={this.props.className}
				width={this.props.size || 24}
				height={this.props.size || 24}
				viewBox="0 0 24 24"
			>
				<path d="M0 0h24v24H0z" fill="none" />
				<path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z" />
			</svg>
		);
	}
}
